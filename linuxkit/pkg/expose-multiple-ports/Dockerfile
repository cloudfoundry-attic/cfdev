FROM linuxkit/alpine:d307c8a386fa3f32cddda9409b9687e191cdd6f1 AS build

RUN apk add --no-cache go musl-dev
ENV GOPATH=/go PATH=$PATH:/go/bin

# See https://github.com/linuxkit/linuxkit/blob/17f5da9fea38f8dfb03e2dd63bf503aad4b84206/tools/alpine/go-compile.sh
# We didn't use go-compile since the linter fails on vpnkit source

COPY *.go /go/src/expose-multiple-ports/
COPY vendor /go/src/expose-multiple-ports/vendor
RUN cd /go/src/expose-multiple-ports && \
    go test && \
    export CGO_ENABLED=0 && \
    go install -buildmode pie -ldflags "-s -w ${ldflags} -extldflags \"-fno-PIC -static\""


FROM scratch
ENTRYPOINT []
CMD []
WORKDIR /
COPY --from=build /go/bin/expose-multiple-ports usr/bin/expose-multiple-ports
