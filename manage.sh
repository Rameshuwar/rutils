#!/bin/bash
# ==============================================================================
# Universal Docker Management Script with Persistent Host Storage
# ==============================================================================

ACTION=$1
CONTAINER_NAME=$2
PORT=$3
IMAGE_NAME=$4
HOST_BASE_DIR=${5:-"/apps/rutils"}

if [[ -z "$ACTION" || -z "$CONTAINER_NAME" || -z "$PORT" || -z "$IMAGE_NAME" ]]; then
    echo "Usage: $0 {start|stop|restart} <container_name> <port> <image_name> [host_base_dir]"
    exit 1
fi

HOST_DATA_DIR="$HOST_BASE_DIR/data"

case "$ACTION" in
    start)
        echo "Starting container: $CONTAINER_NAME"
        
        # 1. Ensure host persistent storage directory exists
        mkdir -p "$HOST_DATA_DIR"
        chmod 755 "$HOST_DATA_DIR"
        
        # 2. Build Port Arguments & Resolve Port Conflicts
        PORT_ARGS=""
        for p in $(echo "$PORT" | tr "," " "); do
            PORT_ARGS="$PORT_ARGS -p $p:$p"
            for conflict_id in $(docker ps -q --filter "publish=$p" 2>/dev/null); do
                if [[ -n "$conflict_id" ]]; then
                    echo "Stopping conflicting container ($conflict_id) occupying port $p..."
                    docker stop "$conflict_id" >/dev/null 2>&1 || true
                fi
            done
        done
        
        # 3. Build Volume Arguments (Persistent data on VPS host /apps/rutils/data)
        VOL_ARGS="-v $HOST_DATA_DIR:/app/data"
        
        # 4. If host has a persistent config.json, mount it as well
        if [[ -f "$HOST_BASE_DIR/config.json" ]]; then
            VOL_ARGS="$VOL_ARGS -v $HOST_BASE_DIR/config.json:/app/config.json"
        fi
        
        echo "Persistent storage bound: $HOST_DATA_DIR -> /app/data"
        docker run -d \
            --name "$CONTAINER_NAME" \
            $PORT_ARGS \
            $VOL_ARGS \
            --restart unless-stopped \
            "$IMAGE_NAME"
        ;;
    stop)
        echo "Stopping container: $CONTAINER_NAME"
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
        docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
        ;;
    restart)
        $0 stop "$CONTAINER_NAME" "$PORT" "$IMAGE_NAME" "$HOST_BASE_DIR"
        $0 start "$CONTAINER_NAME" "$PORT" "$IMAGE_NAME" "$HOST_BASE_DIR"
        ;;
    *)
        echo "Invalid action: $ACTION. Use start, stop, or restart."
        exit 1
esac
