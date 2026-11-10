.PHONY: build help
.DEFAULT_GOAL := build

# ── 交叉编译（证书签发见上级目录的 Makefile） ──────────────
# 钉死 go-1.23.1 镜像：它是最后一批 Ubuntu 20.04（glibc 2.31）基础的 tag，
# 新镜像（24.04/glibc 2.39）编出的动态二进制在老系统（如 TencentOS
# glibc 2.28）上起不来。此 tag 若被上游重推为新基础需重新评估。
XGO_IMAGE ?= techknowlogick/xgo:go-1.23.1
build:
	xgo --targets=darwin/amd64,linux/amd64,linux/arm64 -image $(XGO_IMAGE) -out ./bin/v5-reverse .

help:
	@echo "证书签发见上级 minibridge/ 目录：make certs / make issue"
