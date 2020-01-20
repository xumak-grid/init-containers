# Create a gruop

curl -X POST \
    -u admin:admin123 \
    -H "Content-Type: application/json" \
    -d @examples/createGroup.json \
    http://localhost:8081/service/extdirect


# Create a hosted repository

curl -X POST \
    -u admin:admin123 \
    -H "Content-Type: application/json" \
    -d @examples/createHostedRepo.json
    http://localhost:8081/service/extdirect


# Create a proxy repository

curl -X POST \
    -u admin:admin123 \
    -H "Content-Type: application/json" \
    -d @examples/createProxyRepo.json \
    http://localhost:8081/service/extdirect


# Create a proxy repository without authentication

curl -X POST \
    -u admin:admin123 \
    -H "Content-Type: application/json" \
    -d @examples/createProxyRepoNoAuth.json \
    http://localhost:8081/service/extdirect

# Change admin password
