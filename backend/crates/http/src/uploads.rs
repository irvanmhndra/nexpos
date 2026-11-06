//! `POST /api/v1/uploads`: stores a product image in R2, like the Go `UploadHandler`.
//! Images are validated by their real magic bytes, not the client's content type,
//! and keys are scoped per company.

use axum::{
    Router,
    extract::{Multipart, State, multipart::MultipartRejection},
    http::StatusCode,
    routing::post,
};
use serde::Serialize;

use crate::{
    AppError, AppResult, AppState,
    envelope::{Reply, success},
    extract::AuthUser,
};

pub fn router() -> Router<AppState> {
    Router::new().route("/api/v1/uploads", post(upload))
}

const MAX_UPLOAD_BYTES: usize = 10 << 20;

#[derive(Serialize)]
struct UploadResponse {
    url: String,
    key: String,
}

async fn upload(
    State(state): State<AppState>,
    user: AuthUser,
    multipart: Result<Multipart, MultipartRejection>,
) -> AppResult<Reply<UploadResponse>> {
    let Some(storage) = &state.storage else {
        return Err(AppError::Coded {
            status: StatusCode::INTERNAL_SERVER_ERROR,
            code: "INTERNAL_ERROR",
            message: "object storage is not configured".into(),
        });
    };

    let bytes = read_file(multipart)
        .await
        .ok_or_else(|| AppError::bad_request("file is required"))?;
    if bytes.len() > MAX_UPLOAD_BYTES {
        return Err(AppError::bad_request("file too large (max 10MB)"));
    }
    let (ext, content_type) = sniff_image(&bytes)
        .ok_or_else(|| AppError::bad_request("only JPG, PNG, or WEBP images are allowed"))?;

    let key = format!("products/{}/{}{ext}", user.company_id, uuid::Uuid::new_v4());
    let url = storage.upload(&key, content_type, bytes).await?;
    Ok(success(
        StatusCode::CREATED,
        "Image uploaded",
        UploadResponse { url, key },
    ))
}

/// The bytes of the `file` part, if the request has one.
async fn read_file(multipart: Result<Multipart, MultipartRejection>) -> Option<Vec<u8>> {
    let mut multipart = multipart.ok()?;
    while let Some(field) = multipart.next_field().await.ok()? {
        if field.name() == Some("file") && field.file_name().is_some() {
            return field.bytes().await.ok().map(|b| b.to_vec());
        }
    }
    None
}

/// The extension and content type from a file's magic bytes.
fn sniff_image(head: &[u8]) -> Option<(&'static str, &'static str)> {
    if head.starts_with(&[0xFF, 0xD8, 0xFF]) {
        Some((".jpg", "image/jpeg"))
    } else if head.starts_with(&[0x89, b'P', b'N', b'G', 0x0D, 0x0A, 0x1A, 0x0A]) {
        Some((".png", "image/png"))
    } else if head.len() >= 12 && &head[0..4] == b"RIFF" && &head[8..12] == b"WEBP" {
        Some((".webp", "image/webp"))
    } else {
        None
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn recognizes_images_by_magic_bytes() {
        assert_eq!(
            sniff_image(&[0xFF, 0xD8, 0xFF, 0xE0]),
            Some((".jpg", "image/jpeg"))
        );
        assert_eq!(
            sniff_image(&[0x89, b'P', b'N', b'G', 0x0D, 0x0A, 0x1A, 0x0A, 0]),
            Some((".png", "image/png"))
        );
        assert_eq!(
            sniff_image(b"RIFF\0\0\0\0WEBPVP8 "),
            Some((".webp", "image/webp"))
        );
        assert_eq!(sniff_image(b"GIF89a"), None);
        assert_eq!(sniff_image(b"RIFF"), None);
        assert_eq!(sniff_image(b""), None);
    }
}
