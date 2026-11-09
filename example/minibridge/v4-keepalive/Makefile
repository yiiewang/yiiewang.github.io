.PHONY: all certs ca server client build verify clean help
.DEFAULT_GOAL := certs

# ── 可配置参数（命令行覆盖） ──────────────────────────────
CERT_DIR ?= certs
DAYS     ?= 365
CA_CN    ?= minibridge-demo-ca
SRV_CN   ?= 127.0.0.1
SRV_SAN  ?= IP:127.0.0.1,DNS:localhost
CLI_CN   ?= minibridge-client

# 真握手自检的参数：本地起一个临时服务端，s_client 连上去握一次 mTLS。
# VERIFY_IP 是"客户端视角要校验的地址"——默认 127.0.0.1 与 SRV_SAN 对齐；
# 部署到别的地址时改成 `make verify VERIFY_IP=10.0.0.5` 即可模拟远端校验。
VERIFY_PORT ?= 19443
VERIFY_IP   ?= 127.0.0.1
VERIFY_BIN  := bin/verify-server
VERIFY_LOG  := /tmp/minibridge-verify.log

CA  := $(CERT_DIR)/ca
SRV := $(CERT_DIR)/srv
CLI := $(CERT_DIR)/cli

# ── 一次性生成三张证书 ────────────────────────────────────
all: certs

certs: ca server client
	@echo "证书就绪，位于 $(CERT_DIR)/：ca.crt / srv.crt+srv.key / cli.crt+cli.key"

ca: $(CA).crt
server: $(SRV).crt
client: $(CLI).crt

# ── 自签 CA ───────────────────────────────────────────────
# 私钥用 ecparam 生成：不加密、不交互，也绕开 -nodes / -noenc 的版本差异。
$(CA).key:
	@mkdir -p $(CERT_DIR)
	openssl ecparam -genkey -name prime256v1 -out $(CA).key

$(CA).crt: $(CA).key
	openssl req -new -key $(CA).key -subj "/CN=$(CA_CN)" -out $(CA).csr
	printf '%s\n' \
		'basicConstraints=critical,CA:TRUE' \
		'keyUsage=critical,keyCertSign,cRLSign' \
		'subjectKeyIdentifier=hash' \
		> $(CA).ext
	openssl x509 -req -in $(CA).csr -signkey $(CA).key -days $(DAYS) -sha256 \
		-extfile $(CA).ext -out $(CA).crt
	@rm -f $(CA).csr $(CA).ext
	@echo "CA 已生成: $(CA).crt（对外分发这张）+ $(CA).key（绝不外传）"

# ── 服务端证书 ────────────────────────────────────────────
# SAN 是重点：客户端按 -server 的主机部分校验证书（默认 127.0.0.1），
# Go 早就不认 CN 了，没有 subjectAltName 一定校不过。
$(SRV).key:
	@mkdir -p $(CERT_DIR)
	openssl ecparam -genkey -name prime256v1 -out $(SRV).key

$(SRV).crt: $(SRV).key $(CA).crt
	openssl req -new -key $(SRV).key -subj "/CN=$(SRV_CN)" -out $(SRV).csr
	printf '%s\n' \
		'basicConstraints=critical,CA:FALSE' \
		'keyUsage=critical,digitalSignature,keyEncipherment' \
		'extendedKeyUsage=serverAuth' \
		'subjectAltName=$(SRV_SAN)' \
		> $(SRV).ext
	openssl x509 -req -in $(SRV).csr -CA $(CA).crt -CAkey $(CA).key -CAcreateserial \
		-days $(DAYS) -sha256 -extfile $(SRV).ext -out $(SRV).crt
	@rm -f $(SRV).csr $(SRV).ext
	@echo "服务端证书已生成: $(SRV).crt + $(SRV).key"

# ── 客户端证书 ────────────────────────────────────────────
# 服务端校验客户端时会自动要求 clientAuth，少了这个 EKU 就是
# "certificate specifies an incompatible key usage"。SAN 不需要。
$(CLI).key:
	@mkdir -p $(CERT_DIR)
	openssl ecparam -genkey -name prime256v1 -out $(CLI).key

$(CLI).crt: $(CLI).key $(CA).crt
	openssl req -new -key $(CLI).key -subj "/CN=$(CLI_CN)" -out $(CLI).csr
	printf '%s\n' \
		'basicConstraints=critical,CA:FALSE' \
		'keyUsage=critical,digitalSignature' \
		'extendedKeyUsage=clientAuth' \
		> $(CLI).ext
	openssl x509 -req -in $(CLI).csr -CA $(CA).crt -CAkey $(CA).key -CAcreateserial \
		-days $(DAYS) -sha256 -extfile $(CLI).ext -out $(CLI).crt
	@rm -f $(CLI).csr $(CLI).ext
	@echo "客户端证书已生成: $(CLI).crt + $(CLI).key"

# ── 交叉编译 ──────────────────────────────────────────────
build:
	xgo --targets=darwin/amd64,linux/amd64 -out ./bin/v3-crypto .

# ── 校验：链和用途都对得上 ────────────────────────────────
verify: certs
	openssl verify -CAfile $(CA).crt $(SRV).crt $(CLI).crt
	@echo "--- 服务端 SAN / 用途 ---"
	@openssl x509 -in $(SRV).crt -noout -text | grep -A1 -e 'Subject Alternative Name' -e 'Extended Key Usage'
	@echo "--- 客户端 用途 ---"
	@openssl x509 -in $(CLI).crt -noout -text | grep -A1 'Extended Key Usage'
	@echo "--- 真握手自检 (verify_ip=$(VERIFY_IP)) ---"
	@go build -o $(VERIFY_BIN) .
	@rm -f $(VERIFY_LOG)
	@( $(VERIFY_BIN) server -addr 127.0.0.1:$(VERIFY_PORT) -cert $(SRV).crt -key $(SRV).key -client-ca $(CA).crt >$(VERIFY_LOG) 2>&1 & pid=$$!; \
	sleep 0.7; \
	out=$$(echo Q | timeout 5 openssl s_client -connect 127.0.0.1:$(VERIFY_PORT) -verify_ip $(VERIFY_IP) -CAfile $(CA).crt -cert $(CLI).crt -key $(CLI).key 2>&1); \
	sleep 0.3; kill $$pid 2>/dev/null; \
	if ! echo "$$out" | grep -q 'Verify return code: 0 (ok)'; then \
		echo "$$out" | tail -15; echo "握手失败：客户端没验过服务端证书（先查 SRV_SAN 是否覆盖 VERIFY_IP）"; \
		rm -f $(VERIFY_BIN); exit 1; \
	fi; \
	if ! grep -q 'session closed' $(VERIFY_LOG); then \
		echo "握手失败：服务端没接受客户端证书（对照 -client-ca 与 cli.crt 的签发 CA）"; \
		tail -5 $(VERIFY_LOG); rm -f $(VERIFY_BIN); exit 1; \
	fi; \
	echo "mTLS 真握手通过：链 / SAN / 双向证书全部 OK"; rm -f $(VERIFY_BIN) ) 2>/dev/null

# ── 清理 ──────────────────────────────────────────────────
clean:
	rm -rf $(CERT_DIR)

# ── 帮助 ──────────────────────────────────────────────────
help:
	@echo "用法: make <target> [参数=值 ...]"
	@echo ""
	@echo "生成证书:"
	@echo "  make certs                 一次生成 CA + 服务端 + 客户端（默认目标）"
	@echo "  make ca                    只重建 CA（叶子证书会跟着重签）"
	@echo "  make server / make client  只重建单张叶子证书"
	@echo "  make verify                校验签名链 + 用途，并真握一次 mTLS（治 SAN 不匹配）"
	@echo "  make clean                 删除 $(CERT_DIR)/"
	@echo ""
	@echo "构建:"
	@echo "  make build                 xgo 交叉编译到 ./bin/v3-crypto（需要 xgo）"
	@echo ""
	@echo "可配置参数（默认值）:"
	@echo "  CERT_DIR=$(CERT_DIR)              输出目录"
	@echo "  DAYS=$(DAYS)                    有效期（天）"
	@echo "  CA_CN=$(CA_CN)        CA 的 CommonName"
	@echo "  SRV_CN=$(SRV_CN)               服务端证书 CommonName"
	@echo "  SRV_SAN=$(SRV_SAN)  服务端证书 SAN"
	@echo "  CLI_CN=$(CLI_CN)      客户端证书 CommonName"
	@echo ""
	@echo "示例:"
	@echo "  make certs"
	@echo "  make certs CERT_DIR=/tmp/mb DAYS=30"
	@echo "  make certs SRV_SAN=IP:10.0.0.5,DNS:tunnel.internal"
	@echo ""
	@echo "生成后用起来（两个终端）:"
	@echo "  go run . server -addr 127.0.0.1:9000 -cert $(SRV).crt -key $(SRV).key \\"
	@echo "      -client-ca $(CA).crt -allow 127.0.0.1:8000"
	@echo "  go run . client -listen :8080 -server 127.0.0.1:9000 -target 127.0.0.1:8000 \\"
	@echo "      -ca $(CA).crt -cert $(CLI).crt -key $(CLI).key"
