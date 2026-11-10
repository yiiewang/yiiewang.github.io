.PHONY: build help
.DEFAULT_GOAL := build

# ── 交叉编译（证书签发见上级目录的 Makefile） ──────────────
# 镜像选择原因见 v5-reverse/Makefile：go-1.23.1 是最后一批 glibc 2.31 基础的 tag。
XGO_IMAGE ?= techknowlogick/xgo:go-1.23.1
build:
	xgo --targets=darwin/amd64,linux/amd64 -image $(XGO_IMAGE) -out ./bin/v3-crypto .

help:
	@echo "证书签发 / verify 真握手自检：回到 minibridge/ 目录 make certs / make verify"
