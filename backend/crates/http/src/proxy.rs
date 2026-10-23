//! Strangler facade: requests for routes Rust doesn't serve yet are forwarded
//! to the Go API when `GO_UPSTREAM_URL` is set, so one entry point serves the
//! whole API while route groups move over one by one. Without an upstream,
//! unknown routes get Echo's JSON 404.

use axum::{
    body::Body,
    extract::{Request, State},
    http::{HeaderMap, HeaderName, StatusCode, header},
    response::{IntoResponse, Response},
};

use crate::{AppState, envelope};

/// Hop-by-hop headers (RFC 9110 §7.6.1) never cross a proxy.
const HOP_BY_HOP: [HeaderName; 6] = [
    header::CONNECTION,
    header::PROXY_AUTHENTICATE,
    header::PROXY_AUTHORIZATION,
    header::TE,
    header::TRAILER,
    header::TRANSFER_ENCODING,
];

pub async fn fallback(State(state): State<AppState>, req: Request) -> Response {
    let Some(upstream) = state.upstream else {
        return envelope::failure(StatusCode::NOT_FOUND, "Not Found", "Not Found", None);
    };

    let path = req.uri().path_and_query().map_or("/", |p| p.as_str());
    let url = format!("{}{path}", upstream.base_url);
    let (parts, body) = req.into_parts();

    let mut headers = copy_headers(&parts.headers);
    headers.remove(header::HOST);

    let sent = upstream
        .client
        .request(parts.method, &url)
        .headers(headers)
        .body(reqwest::Body::wrap_stream(body.into_data_stream()))
        .send()
        .await;

    match sent {
        Ok(res) => {
            let mut out = Response::builder().status(res.status());
            if let Some(headers) = out.headers_mut() {
                // This server adds its own CORS and request-id headers; doubled
                // CORS headers make browsers reject the response.
                *headers = copy_headers(res.headers());
                let own: Vec<HeaderName> = headers
                    .keys()
                    .filter(|n| {
                        n.as_str().starts_with("access-control-") || n.as_str() == "x-request-id"
                    })
                    .cloned()
                    .collect();
                for name in own {
                    headers.remove(name);
                }
            }
            out.body(Body::from_stream(res.bytes_stream()))
                .unwrap_or_else(|_| bad_gateway())
        }
        Err(err) => {
            tracing::error!(error = %err, %url, "upstream request failed");
            bad_gateway()
        }
    }
}

fn copy_headers(from: &HeaderMap) -> HeaderMap {
    let mut out = HeaderMap::with_capacity(from.len());
    for (name, value) in from {
        if !HOP_BY_HOP.contains(name) {
            out.append(name.clone(), value.clone());
        }
    }
    out
}

fn bad_gateway() -> Response {
    envelope::failure(StatusCode::BAD_GATEWAY, "Bad Gateway", "Bad Gateway", None).into_response()
}
