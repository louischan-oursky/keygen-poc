FROM golang:1.22.7
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ARG KEYGEN_ACCOUNT_ID
ARG KEYGEN_PRODUCT_ID
RUN make cat KEYGEN_ACCOUNT_ID=${KEYGEN_ACCOUNT_ID} KEYGEN_PRODUCT_ID=${KEYGEN_PRODUCT_ID} && cp ./cat /usr/local/bin/keygen-cat
CMD ["keygen-cat", "--help"]
