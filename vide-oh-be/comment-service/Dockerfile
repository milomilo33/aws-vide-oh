FROM ghcr.io/cargo-lambda/cargo-lambda:latest as builder

RUN rustup install nightly && rustup default nightly

WORKDIR /app

COPY . .

# Clean previous builds and build the project targeting x86_64 (amd64)
RUN cargo clean
RUN cargo +nightly lambda build --release --target x86_64-unknown-linux-gnu

FROM alpine:latest as final
WORKDIR /app

RUN mkdir -p /bin

COPY --from=builder /app/target/lambda/comment-service/bootstrap /app/bin/

# Dummy entrypoint
CMD ["/bin/sh"]