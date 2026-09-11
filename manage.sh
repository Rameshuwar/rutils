#!/bin/bash
# Universal Docker Management Script

ACTION=$1
CONTAINER_NAME=$2
PORT=$3
IMAGE_NAME=$4

if [[ -z "$ACTION" || -z "$CONTAINER_NAME" || -z "$PORT" || -z "$IMAGE_NAME" ]]; then
    echo "Usage: $0 {start|stop|restart} <container_name> <port> <image_name>"
    exit 1
fi

case "$ACTION" in
    start)
        echo "Starting container: $CONTAINER_NAME"
        PORT_ARGS=""
        for p in $(echo "$PORT" | tr "," " "); do
            PORT_ARGS="$PORT_ARGS -p $p:$p"
        done
        docker run -d --name "$CONTAINER_NAME" $PORT_ARGS --restart unless-stopped "$IMAGE_NAME"
        ;;
    stop)
        echo "Stopping container: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
        docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
        ;;
    restart)
        $0 stop "$CONTAINER_NAME" "$PORT" "$IMAGE_NAME"
        $0 start "$CONTAINER_NAME" "$PORT" "$IMAGE_NAME"
        ;;
    *)
        echo "Invalid action: $ACTION. Use start, stop, or restart."
        exit 1
esac
