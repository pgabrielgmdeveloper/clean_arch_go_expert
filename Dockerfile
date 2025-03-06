FROM golang

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
COPY database/migrations /app/database/migrations

RUN go build -o main ./cmd/ordersystem


RUN wget https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz -O migrate.tar.gz && \
    tar -xvzf migrate.tar.gz && \
    mv migrate /usr/local/bin/migrate && \
    rm migrate.tar.gz

EXPOSE 8080 8000 50051

# Comando para executar as migrações e iniciar a aplicação
CMD ["sh", "-c", "migrate -path /app/database/migrations -database 'mysql://root:root@tcp(mysql:3306)/orders?multiStatements=true' up && ./main"]