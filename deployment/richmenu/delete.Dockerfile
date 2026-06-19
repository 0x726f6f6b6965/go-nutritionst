FROM alpine:3.24.1

RUN apk add --no-cache curl

# Run curl via shell to resolve the environment variable
CMD curl -v -X DELETE \
    https://api.line.me/v2/bot/user/all/richmenu \
    -H "Authorization: Bearer ${CHANNEL_ACCESS_TOKEN}"
