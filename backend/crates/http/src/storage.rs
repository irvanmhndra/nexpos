//! Cloudflare R2 over the S3 API: just the two calls nexpos needs, PUT and DELETE,
//! signed with AWS Signature V4. The same public URLs and object settings as the
//! Go `storage.Client`.

use hmac::{Hmac, Mac};
use reqwest::{Client, Method, Url};
use sha2::{Digest, Sha256};

const REGION: &str = "auto";
const SERVICE: &str = "s3";
/// Keys are random and immutable, so the CDN may cache forever.
const CACHE_CONTROL: &str = "public, max-age=31536000, immutable";

#[derive(Debug, Clone)]
pub struct StorageConfig {
    pub bucket: String,
    pub endpoint: String,
    pub access_key_id: String,
    pub secret_key: String,
    pub public_base_url: String,
}

impl StorageConfig {
    /// `None` unless the bucket, endpoint, and key are all set, like Go's `storage.New`.
    pub fn complete(
        bucket: String,
        endpoint: String,
        access_key_id: String,
        secret_key: String,
        public_base_url: String,
    ) -> Option<Self> {
        if bucket.is_empty() || endpoint.is_empty() || access_key_id.is_empty() {
            return None;
        }
        Some(Self {
            bucket,
            endpoint,
            access_key_id,
            secret_key,
            public_base_url,
        })
    }
}

pub struct Storage {
    config: StorageConfig,
    endpoint: Url,
    public_base_url: String,
    http: Client,
}

impl Storage {
    pub fn new(config: StorageConfig) -> anyhow::Result<Self> {
        let endpoint = Url::parse(&config.endpoint)?;
        let public_base_url = config.public_base_url.trim_end_matches('/').to_owned();
        Ok(Self {
            config,
            endpoint,
            public_base_url,
            http: Client::new(),
        })
    }

    /// Stores an object and returns its public URL.
    pub async fn upload(
        &self,
        key: &str,
        content_type: &str,
        body: Vec<u8>,
    ) -> anyhow::Result<String> {
        self.send(
            Method::PUT,
            key,
            &[
                ("cache-control", CACHE_CONTROL),
                ("content-type", content_type),
            ],
            body,
        )
        .await?;
        Ok(self.public_url(key))
    }

    pub async fn delete(&self, key: &str) -> anyhow::Result<()> {
        self.send(Method::DELETE, key, &[], Vec::new()).await
    }

    pub fn public_url(&self, key: &str) -> String {
        format!("{}/{}", self.public_base_url, key.trim_start_matches('/'))
    }

    /// The object key of a URL this store owns, or `None` for an empty or foreign URL.
    pub fn key_from_url(&self, url: &str) -> Option<String> {
        if url.is_empty() || self.public_base_url.is_empty() {
            return None;
        }
        url.strip_prefix(&format!("{}/", self.public_base_url))
            .map(str::to_owned)
    }

    async fn send(
        &self,
        method: Method,
        key: &str,
        extra_headers: &[(&str, &str)],
        body: Vec<u8>,
    ) -> anyhow::Result<()> {
        let path = format!("/{}/{}", uri_encode(&self.config.bucket), uri_encode(key));
        let mut url = self.endpoint.clone();
        url.set_path(&path);
        let host = match url.port() {
            Some(port) => format!("{}:{port}", url.host_str().unwrap_or_default()),
            None => url.host_str().unwrap_or_default().to_owned(),
        };

        let payload_hash = hex::encode(Sha256::digest(&body));
        let amz_date = chrono::Utc::now().format("%Y%m%dT%H%M%SZ").to_string();
        let mut signed: Vec<(&str, &str)> = vec![
            ("host", &host),
            ("x-amz-content-sha256", &payload_hash),
            ("x-amz-date", &amz_date),
        ];
        signed.extend_from_slice(extra_headers);
        let authorization = authorization(
            method.as_str(),
            &path,
            &signed,
            &payload_hash,
            &amz_date,
            REGION,
            &self.config.access_key_id,
            &self.config.secret_key,
        );

        let mut request = self
            .http
            .request(method, url)
            .header("authorization", authorization);
        for (name, value) in signed.iter().filter(|(name, _)| *name != "host") {
            request = request.header(*name, *value);
        }
        let response = request.body(body).send().await?;
        if !response.status().is_success() {
            let status = response.status();
            let detail = response.text().await.unwrap_or_default();
            anyhow::bail!("object storage answered {status}: {detail}");
        }
        Ok(())
    }
}

/// The `Authorization` header for AWS Signature V4 with no query string.
/// Header names must be lowercase.
#[allow(clippy::too_many_arguments)]
fn authorization(
    method: &str,
    path: &str,
    headers: &[(&str, &str)],
    payload_hash: &str,
    amz_date: &str,
    region: &str,
    access_key_id: &str,
    secret_key: &str,
) -> String {
    let mut sorted: Vec<&(&str, &str)> = headers.iter().collect();
    sorted.sort_by_key(|(name, _)| *name);
    let canonical_headers: String = sorted
        .iter()
        .map(|(name, value)| format!("{name}:{}\n", value.trim()))
        .collect();
    let signed_headers = sorted
        .iter()
        .map(|(name, _)| *name)
        .collect::<Vec<_>>()
        .join(";");
    let canonical_request =
        format!("{method}\n{path}\n\n{canonical_headers}\n{signed_headers}\n{payload_hash}");

    let date = &amz_date[..8];
    let scope = format!("{date}/{region}/{SERVICE}/aws4_request");
    let to_sign = format!(
        "AWS4-HMAC-SHA256\n{amz_date}\n{scope}\n{}",
        hex::encode(Sha256::digest(canonical_request.as_bytes()))
    );

    let key = hmac(format!("AWS4{secret_key}").as_bytes(), date.as_bytes());
    let key = hmac(&key, region.as_bytes());
    let key = hmac(&key, SERVICE.as_bytes());
    let key = hmac(&key, b"aws4_request");
    let signature = hex::encode(hmac(&key, to_sign.as_bytes()));

    format!(
        "AWS4-HMAC-SHA256 Credential={access_key_id}/{scope},SignedHeaders={signed_headers},Signature={signature}"
    )
}

fn hmac(key: &[u8], data: &[u8]) -> Vec<u8> {
    let mut mac = Hmac::<Sha256>::new_from_slice(key).expect("HMAC accepts any key length");
    mac.update(data);
    mac.finalize().into_bytes().to_vec()
}

/// Percent-encodes a path, keeping `/` and the characters S3 leaves unreserved.
fn uri_encode(path: &str) -> String {
    let mut out = String::with_capacity(path.len());
    for byte in path.bytes() {
        match byte {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' | b'/' => {
                out.push(byte as char)
            }
            _ => out.push_str(&format!("%{byte:02X}")),
        }
    }
    out
}

#[cfg(test)]
mod tests {
    use super::*;

    const SECRET: &str = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY";
    const EMPTY: &str = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

    /// The worked "GET Object" example in AWS's Signature V4 documentation.
    #[test]
    fn signature_matches_the_aws_documentation_example() {
        let header = authorization(
            "GET",
            "/test.txt",
            &[
                ("host", "examplebucket.s3.amazonaws.com"),
                ("range", "bytes=0-9"),
                ("x-amz-content-sha256", EMPTY),
                ("x-amz-date", "20130524T000000Z"),
            ],
            EMPTY,
            "20130524T000000Z",
            "us-east-1",
            "AKIAIOSFODNN7EXAMPLE",
            SECRET,
        );
        assert_eq!(
            header,
            "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20130524/us-east-1/s3/aws4_request,\
             SignedHeaders=host;range;x-amz-content-sha256;x-amz-date,\
             Signature=f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41"
        );
    }

    #[test]
    fn config_needs_bucket_endpoint_and_key() {
        let s = String::new;
        assert!(StorageConfig::complete(s(), "e".into(), "k".into(), s(), s()).is_none());
        assert!(StorageConfig::complete("b".into(), s(), "k".into(), s(), s()).is_none());
        assert!(StorageConfig::complete("b".into(), "e".into(), s(), s(), s()).is_none());
        assert!(StorageConfig::complete("b".into(), "e".into(), "k".into(), s(), s()).is_some());
    }

    #[test]
    fn urls_map_to_keys_only_for_this_store() {
        let storage = Storage::new(StorageConfig {
            bucket: "b".into(),
            endpoint: "https://acct.r2.cloudflarestorage.com".into(),
            access_key_id: "k".into(),
            secret_key: "s".into(),
            public_base_url: "https://cdn.example.com/".into(),
        })
        .unwrap();
        assert_eq!(
            storage.public_url("/products/1/a.png"),
            "https://cdn.example.com/products/1/a.png"
        );
        assert_eq!(
            storage
                .key_from_url("https://cdn.example.com/products/1/a.png")
                .as_deref(),
            Some("products/1/a.png")
        );
        assert_eq!(storage.key_from_url("https://elsewhere.com/a.png"), None);
        assert_eq!(storage.key_from_url(""), None);
    }

    #[test]
    fn encodes_paths_like_s3() {
        assert_eq!(
            uri_encode("products/1/a b+c.png"),
            "products/1/a%20b%2Bc.png"
        );
    }
}
