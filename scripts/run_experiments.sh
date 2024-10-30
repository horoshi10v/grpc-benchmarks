#!/bin/bash

# Параметри експериментів
SERVER_ADDRESS="localhost:50052"
EXPERIMENTS_DIR="../results"
LOG_DIR="$EXPERIMENTS_DIR/logs"
METRICS_DIR="$EXPERIMENTS_DIR/metrics"

# Створення директорій для збереження результатів
mkdir -p $LOG_DIR
mkdir -p $METRICS_DIR

# Функція для генерації повідомлень різного розміру
generate_message() {
    SIZE=$1
    if [ "$SIZE" == "small" ]; then
        MESSAGE="Hello"
    elif [ "$SIZE" == "medium" ]; then
        MESSAGE=$(head /dev/urandom | LC_CTYPE=C tr -dc 'A-Za-z0-9' | head -c 1024)
    elif [ "$SIZE" == "large" ]; then
        MESSAGE=$(head /dev/urandom | LC_CTYPE=C tr -dc 'A-Za-z0-9' | head -c 1048576)
    else
        MESSAGE="Default Message"
    fi
    echo "$MESSAGE"
}

# Функція для запуску сервера
start_server() {
    echo "Starting server..."
    # Запуск сервера у фоновому режимі та збереження PID
    go run ../server/*.go > $LOG_DIR/server.log 2>&1 &
    SERVER_PID=$!
    echo "Server PID: $SERVER_PID"
    # Чекаємо, поки сервер запуститься
    sleep 2
}

# Функція для зупинки сервера
stop_server() {
    echo "Stopping server..."
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null
}

# Функція для запуску клієнта з певними параметрами
run_client() {
    MODE=$1
    ACTION=$2
    TOPIC_OR_QUEUE=$3
    MESSAGE_SIZE=$4
    OUTPUT_FILE=$5
    NUM_REQUESTS=$6
    CONCURRENCY=$7

    echo "Running client in mode: $MODE, action: $ACTION, topic/queue: $TOPIC_OR_QUEUE, message size: $MESSAGE_SIZE, requests: $NUM_REQUESTS, concurrency: $CONCURRENCY"

    MESSAGE=$(generate_message $MESSAGE_SIZE)

    if [ "$MODE" == "sync" ] || [ "$MODE" == "async" ]; then
        # Запускаємо колектор метрик
        go run collect_metrics.go --mode=$MODE --requests=$NUM_REQUESTS --concurrency=$CONCURRENCY --message="$MESSAGE" --output=$OUTPUT_FILE > $LOG_DIR/${MODE}_client.log 2>&1
    elif [ "$MODE" == "pubsub" ] || [ "$MODE" == "broker" ]; then
        if [ "$ACTION" == "publish" ]; then
            for ((i=1; i<=$NUM_REQUESTS; i++)); do
                go run ../client/*.go --mode=$MODE --action=publish --topic=$TOPIC_OR_QUEUE --message="$MESSAGE $i" >> $LOG_DIR/${MODE}_publisher.log 2>&1
            done
        elif [ "$ACTION" == "subscribe" ]; then
            go run collect_metrics.go --mode=$MODE --action=subscribe --topic=$TOPIC_OR_QUEUE --output=$OUTPUT_FILE > $LOG_DIR/${MODE}_subscriber.log 2>&1 &
            CLIENT_PID=$!
            # Даємо час на запуск підписника
            sleep 2
        fi
    fi
}

# Запуск серверних експериментів
start_server

# Визначаємо рівні навантаження
LOAD_LEVELS=("low" "medium" "high")
REQUESTS=("10" "50" "100")
CONCURRENCIES=("1" "10" "50")
MESSAGE_SIZES=("small" "medium" "large")

for SIZE in "${MESSAGE_SIZES[@]}"; do
    for i in "${!LOAD_LEVELS[@]}"; do
        LOAD="${LOAD_LEVELS[$i]}"
        NUM_REQUESTS=${REQUESTS[$i]}
        CONCURRENCY=${CONCURRENCIES[$i]}

        echo "Starting experiments with $LOAD load and $SIZE messages: $NUM_REQUESTS requests, $CONCURRENCY concurrency"

        # Синхронний метод
        run_client "sync" "" "" $SIZE "$METRICS_DIR/sync_${LOAD}_${SIZE}_metrics.csv" $NUM_REQUESTS $CONCURRENCY

        # Асинхронний метод
        run_client "async" "" "" $SIZE "$METRICS_DIR/async_${LOAD}_${SIZE}_metrics.csv" $NUM_REQUESTS $CONCURRENCY

        # Pub/Sub метод
        run_client "pubsub" "subscribe" "test_topic" "" "$METRICS_DIR/pubsub_${LOAD}_${SIZE}_metrics.csv" $NUM_REQUESTS $CONCURRENCY
        # Даємо час на запуск підписника
        sleep 2
        run_client "pubsub" "publish" "test_topic" $SIZE "" $NUM_REQUESTS $CONCURRENCY
        # Зупиняємо підписника
        if [ ! -z "$CLIENT_PID" ]; then
            kill $CLIENT_PID
            wait $CLIENT_PID 2>/dev/null
        fi

        # Broker метод
        run_client "broker" "subscribe" "test_queue" "" "$METRICS_DIR/broker_${LOAD}_${SIZE}_metrics.csv" $NUM_REQUESTS $CONCURRENCY
        # Даємо час на запуск підписника
        sleep 2
        run_client "broker" "publish" "test_queue" $SIZE "" $NUM_REQUESTS $CONCURRENCY
        # Зупиняємо підписника
        if [ ! -z "$CLIENT_PID" ]; then
            kill $CLIENT_PID
            wait $CLIENT_PID 2>/dev/null
        fi
    done
done

# Зупинка серверних процесів
stop_server

echo "Experiments completed. Logs and metrics are saved in $EXPERIMENTS_DIR."
