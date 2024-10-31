# Stage 1: Build golang dependencies and binaries
FROM ubuntu:24.10 AS build

ARG VM_DEBUG

RUN apt-get -qq update && \
    apt-get -qq install curl build-essential git golang upx-ucl libjemalloc-dev libjemalloc2 -y
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -q -y

WORKDIR /plugin

COPY . .

RUN git clone https://github.com/NethermindEth/juno.git && \
    cd juno && \
    git checkout pitchlake/plugin-sequencer

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
COPY --from=build /plugin/juno/genesis ./genesis

# Run Juno with the plugin
CMD ["bash", "-c", "./build/juno --plugin-path myplugin.so --http --http-port=6060 --http-host=0.0.0.0 --db-path=../seq-db --log-level=debug --seq-enable --seq-block-time=1 --network sequencer --seq-genesis-file genesis/genesis_prefund_accounts.json --rpc-call-max-steps=4123000 --rpc-cors-enable"]