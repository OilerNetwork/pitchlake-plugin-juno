# Stage 1: Build golang dependencies and binaries
FROM ubuntu:24.10 AS build

ARG VM_DEBUG


RUN apt-get -qq update && \
    apt-get -qq install curl build-essential git golang upx-ucl libjemalloc-dev libjemalloc2 -y
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -q -y

WORKDIR /plugin

# Copy source code
COPY . .

# Build the project
RUN bash -c 'cd juno && source ~/.cargo/env && VM_DEBUG=${VM_DEBUG} make juno'

RUN pwd
RUN ls

# Then build the plugin
RUN make build

# Stage 2: Run Juno with the plugin
FROM ubuntu:24.10

# Install necessary runtime dependencies
RUN apt-get -qq update && \
    apt-get -qq install libjemalloc2 -y && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the Juno binary and the plugin from the build stage
COPY --from=build /plugin/juno/build/juno ./build/
COPY --from=build /plugin/myplugin.so ./

# Run Juno with the plugin
CMD ["bash", "-c", "./build/juno --plugin-path myplugin.so"]