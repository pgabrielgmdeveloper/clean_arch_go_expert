package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"

	"log"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pgabrielgmdeveloper/clean_arch_go_expert/configs"
	"github.com/pgabrielgmdeveloper/clean_arch_go_expert/internal/event/handler"
	"github.com/pgabrielgmdeveloper/clean_arch_go_expert/internal/infra/graph"
	"github.com/pgabrielgmdeveloper/clean_arch_go_expert/internal/infra/grpc/pb"
	"github.com/pgabrielgmdeveloper/clean_arch_go_expert/internal/infra/grpc/service"
	"github.com/pgabrielgmdeveloper/clean_arch_go_expert/internal/infra/web/webserver"
	"github.com/pgabrielgmdeveloper/clean_arch_go_expert/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		fmt.Println("Error connecting to database")
		panic(err)
	}
	defer db.Close()

	migrationsPath := "file://../../database/migrations"

	// Cria uma nova instância do Migrate
	m, err := migrate.New(migrationsPath, "mysql://root:root@tcp(localhost:3306)/orders")
	if err != nil {
		log.Fatalf("Erro ao criar instância do Migrate: %v", err)
	}

	// Executa as migrações "up"
	if err := m.Up(); err != nil {
		log.Fatalf("Erro ao executar migrações: %v", err)
	}

	log.Println("Migrações aplicadas com sucesso!")

	rabbitMQChannel := getRabbitMQChannel()

	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})

	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)
	getOrderUseCase := NewGetOrderUseCase(db)

	webserver := webserver.NewWebServer(configs.WebServerPort)
	webOrderHandler := NewWebOrderHandler(db, eventDispatcher)
	webserver.AddHandler("/order", webOrderHandler.Handle)
	fmt.Println("Starting web server on port", configs.WebServerPort)
	go webserver.Start()

	grpcServer := grpc.NewServer()
	createOrderService := service.NewOrderService(*createOrderUseCase, *getOrderUseCase)
	pb.RegisterOrderServiceServer(grpcServer, createOrderService)
	reflection.Register(grpcServer)

	fmt.Println("Starting gRPC server on port", configs.GRPCServerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", configs.GRPCServerPort))
	if err != nil {
		panic(err)
	}
	go grpcServer.Serve(lis)

	srv := graphql_handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase: *createOrderUseCase,
		GetOrderUseCase:    *getOrderUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Println("Starting GraphQL server on port", configs.GraphQLServerPort)
	http.ListenAndServe(":"+configs.GraphQLServerPort, nil)
}

func getRabbitMQChannel() *amqp.Channel {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}
