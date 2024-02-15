# run-service:
# 	docker-compose up -d

# test:
# 	go run cmd/main.go

commit:
	@read -p "Select commit type (feat/fix): " type && \
	if [ "$$type" != "feat" ] && [ "$$type" != "fix" ]; then \
		echo "Invalid commit type. Please enter 'feat' or 'fix'." && exit 1; \
	fi && \
	read -p "Enter a short commit message: " message && \
	git commit -m "$$type: $$message"