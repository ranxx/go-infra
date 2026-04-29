.PHONY: tag release new-tag

release:
	@set -e; \
	echo "Fetching latest code & tags..."; \
	git fetch --tags; \
	git pull --ff-only; \
	echo "Adding changes..."; \
	git add .; \
	git commit -m "auto" || echo "No changes to commit"; \
	git push; \
	echo "Getting latest version tag..."; \
	latest_tag=$$(git tag --list 'v*' --sort=-v:refname | head -n 1); \
	[ -z "$$latest_tag" ] && latest_tag="v0.0.0"; \
	echo "Latest tag: $$latest_tag"; \
	if [ -z "$$latest_tag" ]; then \
		major=0; minor=0; patch=0; \
	else \
		major=$$(echo "$$latest_tag" | cut -d. -f1 | tr -d 'v'); \
		minor=$$(echo "$$latest_tag" | cut -d. -f2); \
		patch=$$(echo "$$latest_tag" | cut -d. -f3); \
	fi; \
	patch=$$((patch + 1)); \
	new_tag="v$$major.$$minor.$$patch"; \
	echo "New tag: $$new_tag"; \
	git tag -a "$$new_tag" -m "release $$new_tag"; \
	git push origin "$$new_tag"; \
	echo "Released $$new_tag"

new-tag: release
tag: release
