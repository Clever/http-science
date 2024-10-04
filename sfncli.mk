# This is the default sfncli Makefile.
# Please do not alter this file directly.
SFNCLI_MK_VERSION := 0.1.2
SHELL := /bin/bash
SYSTEM := $(shell uname -a | cut -d" " -f1 | tr '[:upper:]' '[:lower:]')
SFNCLI_INSTALLED := $(shell [[ -e "bin/sfncli" ]] && bin/sfncli --version)
# AUTH_HEADER is used to help avoid github ratelimiting
AUTH_HEADER = $(shell [[ ! -z "${GITHUB_API_TOKEN}" ]] && echo "Authorization: token $(GITHUB_API_TOKEN)")
SFNCLI_LATEST = $(shell \
	curl --retry 5 -f -s --header "$(AUTH_HEADER)" \
		https://api.github.com/repos/Clever/sfncli/releases/latest | \
	grep tag_name | \
	cut -d\" -f4)

.PHONY: bin/sfncli sfncli-update-makefile ensure-sfncli-version-set ensure-curl-installed

ensure-sfncli-version-set:
	@ if [[ "$(SFNCLI_VERSION)" = "" ]]; then \
		echo "SFNCLI_VERSION not set in Makefile - Suggest setting 'SFNCLI_VERSION := latest'"; \
		exit 1; \
	fi

ensure-curl-installed:
	@command -v curl >/dev/null 2>&1 || { echo >&2 "curl not installed. Please install curl."; exit 1; }

bin/sfncli: ensure-sfncli-version-set ensure-curl-installed
	@mkdir -p bin
	$(eval SFNCLI_VERSION := $(if $(filter latest,$(SFNCLI_VERSION)),$(SFNCLI_LATEST),$(SFNCLI_VERSION)))
	@echo "Checking for sfncli updates..."
	@# AUTH_HEADER not added to curl command below because it doesn't play well with redirects
	@if [[ "$(SFNCLI_VERSION)" == "$(SFNCLI_INSTALLED)" ]]; then \
		{ [[ -z "$(SFNCLI_INSTALLED)" ]] && \
			{ echo "❌  Error: Failed to download sfncli.  Try setting GITHUB_API_TOKEN"; exit 1; } || \
			{ echo "Using latest sfncli version $(SFNCLI_VERSION)"; } \
		} \
	else \
		echo "Updating sfncli..."; \
		curl --retry 5 --fail --max-time 30 -o bin/sfncli -sL https://github.com/Clever/sfncli/releases/download/$(SFNCLI_VERSION)/sfncli-$(SFNCLI_VERSION)-$(SYSTEM)-amd64 && \
		chmod +x bin/sfncli && \
		echo "Successfully updated sfncli to $(SFNCLI_VERSION)" || \
		{ [[ -z "$(SFNCLI_INSTALLED)" ]] && \
			{ echo "❌  Error: Failed to update sfncli"; exit 1; } || \
			{ echo "⚠️  Warning: Failed to update sfncli using pre-existing version"; } \
		} \
	;fi

sfncli-update-makefile: ensure-curl-installed
	@curl -o /tmp/sfncli.mk -sL https://raw.githubusercontent.com/Clever/sfncli/master/make/sfncli.mk
	@if ! grep -q $(SFNCLI_MK_VERSION) /tmp/sfncli.mk; then cp /tmp/sfncli.mk sfncli.mk && echo "sfncli.mk updated"; else echo "sfncli.mk is up-to-date"; fi
