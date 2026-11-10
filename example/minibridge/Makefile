.PHONY: all certs ca server client issue clean help
.DEFAULT_GOAL := certs

# ── 可配置参数（命令行覆盖） ──────────────────────────────
CERT_DIR ?= certs
# 默认 90 天：本项目没有 CRL/OCSP 吊销机制，证书被盗只能等它过期——
# 短周期是最实际的风险窗口补偿。
DAYS     ?= 90
CA_CN    ?= minibridge-demo-ca
SRV_CN   ?= 127.0.0.1
SRV_SAN  ?= IP:127.0.0.1,DNS:localhost
CLI_CN   ?= minibridge-client
# 客户端证书绑定的来源 IP，逗号分隔多个（如 CLI_IP=10.0.0.5,10.0.0.6）。
# 留空不绑定；绑定后整套证书拷到别的机器（来源 IP 不在 SAN 里）就连不上。
comma := ,
CLI_IP  ?=
CLI_SAN := $(if $(CLI_IP),IP:$(subst $(comma),$(comma)IP:,$(CLI_IP)))
# make issue 的新客户端名：产出 $(CERT_DIR)/<name>.crt + <name>.key
CLI_NAME ?=

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
# "certificate specifies an incompatible key usage"。
# SAN 在这里不是给标准校验用的，是给服务端的 IP 绑定闸门用的：
# 证书里写了 IP SAN，服务端握手后就比对来源地址，不匹配即断开。
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
	$(if $(CLI_SAN),printf 'subjectAltName=%s\n' '$(CLI_SAN)' >> $(CLI).ext)
	openssl x509 -req -in $(CLI).csr -CA $(CA).crt -CAkey $(CA).key -CAcreateserial \
		-days $(DAYS) -sha256 -extfile $(CLI).ext -out $(CLI).crt
	@rm -f $(CLI).csr $(CLI).ext
	@echo "客户端证书已生成: $(CLI).crt + $(CLI).key$(if $(CLI_SAN),（绑定来源 IP: $(CLI_IP)）)"

# ── 新增客户端证书（不动 CA 与已有证书） ──────────────────
# 用现有 CA 给新客户端签一张：服务端 -client-ca 是同一张 CA，即刻可用。
# 用法: make issue CLI_NAME=<名字> [CLI_IP=<ip,ip>]
issue:
	@test -n "$(CLI_NAME)" || { echo "用法: make issue CLI_NAME=<名字> [CLI_IP=<ip>]"; exit 1; }
	@test -f $(CA).crt || { echo "没有 CA（先 make certs）: $(CA).crt"; exit 1; }
	@mkdir -p $(CERT_DIR)
	openssl ecparam -genkey -name prime256v1 -out $(CERT_DIR)/$(CLI_NAME).key
	openssl req -new -key $(CERT_DIR)/$(CLI_NAME).key -subj "/CN=$(CLI_NAME)" -out $(CERT_DIR)/$(CLI_NAME).csr
	printf '%s\n' \
		'basicConstraints=critical,CA:FALSE' \
		'keyUsage=critical,digitalSignature' \
		'extendedKeyUsage=clientAuth' \
		> $(CERT_DIR)/$(CLI_NAME).ext
	$(if $(CLI_IP),printf 'subjectAltName=IP:%s\n' '$(subst $(comma),$(comma)IP:,$(CLI_IP))' >> $(CERT_DIR)/$(CLI_NAME).ext)
	openssl x509 -req -in $(CERT_DIR)/$(CLI_NAME).csr -CA $(CA).crt -CAkey $(CA).key -CAcreateserial \
		-days $(DAYS) -sha256 -extfile $(CERT_DIR)/$(CLI_NAME).ext -out $(CERT_DIR)/$(CLI_NAME).crt
	@rm -f $(CERT_DIR)/$(CLI_NAME).csr $(CERT_DIR)/$(CLI_NAME).ext
	@echo "新客户端证书已签发: $(CERT_DIR)/$(CLI_NAME).crt + $(CERT_DIR)/$(CLI_NAME).key$(if $(CLI_IP),（绑定来源 IP: $(CLI_IP)）)，服务端 -client-ca 不用改"

# ── 清理 ──────────────────────────────────────────────────
clean:
	rm -rf $(CERT_DIR) bin

# ── 帮助 ──────────────────────────────────────────────────
help:
	@echo "用法: make <target> [参数=值 ...]"
	@echo ""
	@echo "生成证书（三个版本共用，产物在本目录 $(CERT_DIR)/）:"
	@echo "  make certs                 一次生成 CA + 服务端 + 客户端（默认目标）"
	@echo "  make ca                    只重建 CA（叶子证书会跟着重签）"
	@echo "  make server / make client  只重建单张叶子证书"
	@echo "  make issue CLI_NAME=名     用现有 CA 新签一张客户端证书（不动已有证书）"
	@echo "  make clean                 删除 $(CERT_DIR)/ 与 bin/"
	@echo ""
	@echo "构建（在各版本目录内，或直接 go build ./v5-reverse）:"
	@echo "  cd v3-crypto && make build"
	@echo "  cd v4-keepalive && make build"
	@echo "  cd v5-reverse && make build"
	@echo ""
	@echo "可配置参数（默认值）:"
	@echo "  CERT_DIR=$(CERT_DIR)              输出目录"
	@echo "  DAYS=$(DAYS)                    有效期（天）"
	@echo "  CA_CN=$(CA_CN)        CA 的 CommonName"
	@echo "  SRV_CN=$(SRV_CN)               服务端证书 CommonName"
	@echo "  SRV_SAN=$(SRV_SAN)  服务端证书 SAN"
	@echo "  CLI_CN=$(CLI_CN)      客户端证书 CommonName"
	@echo "  CLI_IP=$(CLI_IP)                客户端证书绑定的来源 IP（空=不绑定）"
	@echo ""
	@echo "示例:"
	@echo "  make certs"
	@echo "  make certs CERT_DIR=/tmp/mb DAYS=30"
	@echo "  make certs SRV_SAN=IP:10.0.0.5,DNS:tunnel.internal"
	@echo "  make certs CLI_IP=10.0.0.5     # 客户端证书绑定来源 IP，拷走即失效"
	@echo ""
	@echo "生成后用起来（两个终端，在 minibridge/ 下）:"
	@echo "  go run ./v5-reverse server -addr 127.0.0.1:9000 -cert $(SRV).crt -key $(SRV).key \\"
	@echo "      -client-ca $(CA).crt -allow 127.0.0.1:8000"
	@echo "  go run ./v5-reverse client -listen :8080 -server 127.0.0.1:9000 -target 127.0.0.1:8000 \\"
	@echo "      -ca $(CA).crt -cert $(CLI).crt -key $(CLI).key"
