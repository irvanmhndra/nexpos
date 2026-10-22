# nexpos Rust API. cargo-chef caches the dependency build, so a code-only
# change rebuilds just the workspace crates.
FROM rust:1.95-bookworm AS chef
RUN cargo install cargo-chef --locked
WORKDIR /app

FROM chef AS planner
COPY . .
RUN cargo chef prepare --recipe-path recipe.json

FROM chef AS builder
COPY --from=planner /app/recipe.json recipe.json
RUN cargo chef cook --release --recipe-path recipe.json
COPY . .
RUN cargo build --release --bin nexpos-server

FROM gcr.io/distroless/cc-debian12:nonroot
COPY --from=builder /app/target/release/nexpos-server /nexpos-server
EXPOSE 8090
ENTRYPOINT ["/nexpos-server"]
