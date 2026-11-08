FROM alpine:3.20
WORKDIR /app
COPY webhook-receiver/receiver /app/receiver
RUN chmod +x /app/receiver
EXPOSE 5001
ENTRYPOINT ["/app/receiver"]
