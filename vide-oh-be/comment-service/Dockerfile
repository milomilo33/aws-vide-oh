# # Use the official Cargo Lambda Docker image as the base image
# FROM ghcr.io/cargo-lambda/cargo-lambda:latest as builder

# # # Install cross-compilation dependencies
# # RUN apt-get update && apt-get install -y \
# #     gcc-aarch64-linux-gnu \
# #     libc6-dev-arm64-cross \
# #     # libssl-dev \
# #     pkg-config \
# #     build-essential \
# #     curl

# # # Set the target architecture for cross-compilation
# # ENV CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER=aarch64-linux-gnu-gcc
# # ENV OPENSSL_DIR=/usr/lib/aarch64-linux-gnu
# # ENV OPENSSL_INCLUDE_DIR=/usr/include

# # Create a new directory for your project
# WORKDIR /app

# # Copy the project files
# COPY . .

# # RUN ls -la /usr/lib/aarch64-linux-gnu
# # RUN ls -la /usr/include/openssl

# # Build the project targeting arm64
# RUN cargo lambda build --release --arm64

# # Final stage: minimal image for the binary
# FROM debian:buster-slim

# # Copy the built binaries from the builder stage
# COPY --from=builder /app/target/aarch64-unknown-linux-gnu/release /app/

# # Set the entry point
# CMD ["/app/comment-service"]


# FROM ghcr.io/cargo-lambda/cargo-lambda:latest as builder

# # Install cross-compilation dependencies
# RUN apt-get update && apt-get install -y \
#     gcc-aarch64-linux-gnu \
#     libc6-dev-arm64-cross \
#     pkg-config \
#     libssl-dev \
#     libudev-dev \
#     build-essential \
#     curl

# # Verify installation of tools and paths
# RUN openssl version && \
#     pkg-config --version && \
#     aarch64-linux-gnu-gcc --version && \
#     echo "OpenSSL lib path: $(pkg-config --variable=libdir openssl)" && \
#     echo "OpenSSL include path: $(pkg-config --variable=includedir openssl)" && \
#     whereis openssl && \
#     whereis aarch64-linux-gnu-gcc

# # Install Rust nightly
# RUN rustup install nightly && rustup default nightly

# # Create a new directory for your project
# WORKDIR /app

# # Copy the project files
# COPY . .

# # Build the project targeting arm64
# RUN cargo clean
# RUN cargo +nightly lambda build --release --arm64

# # Final stage: minimal image for the binary
# FROM debian:buster-slim

# # Copy the built binaries from the builder stage
# COPY --from=builder /app/target/aarch64-unknown-linux-gnu/release /app/

# # Set the entry point
# CMD ["/app/comment-service"]

# Use the official Cargo Lambda Docker image as the base image
FROM ghcr.io/cargo-lambda/cargo-lambda:latest as builder

# Install cross-compilation dependencies for x86_64 (amd64)
RUN apt-get update && apt-get install -y \
    gcc-x86-64-linux-gnu \
    libc6-dev-amd64-cross \
    pkg-config \
    libssl-dev \
    libudev-dev \
    build-essential \
    curl

# Install Rust nightly (if needed by your project)
RUN rustup install nightly && rustup default nightly

# Set the target architecture for cross-compilation
ENV CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER=x86_64-linux-gnu-gcc
ENV PKG_CONFIG_ALLOW_CROSS=1
ENV PKG_CONFIG_SYSROOT_DIR=/usr/x86_64-linux-gnu
ENV PKG_CONFIG_PATH=/usr/lib/x86_64-linux-gnu/pkgconfig

# Create a new directory for your project
WORKDIR /app

# Copy the project files
COPY . .

# Clean previous builds and build the project targeting x86_64 (amd64)
RUN cargo clean
RUN cargo +nightly lambda build --release --target x86_64-unknown-linux-gnu

# Final stage: minimal image for the binary
FROM debian:buster-slim

# Copy the built binaries from the builder stage
COPY --from=builder /app/target/x86_64-unknown-linux-gnu/release /app/

# Set the entry point
CMD ["/app/comment-service"]


