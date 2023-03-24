package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AdiKhoironHasan/bookservices/cmd"
	"github.com/AdiKhoironHasan/bookservices/config"
	"github.com/AdiKhoironHasan/bookservices/domain/service"
	"github.com/AdiKhoironHasan/bookservices/grpc/client"
	"github.com/AdiKhoironHasan/bookservices/infrastructure/persistence"
	"github.com/AdiKhoironHasan/bookservices/rest"
	"github.com/AdiKhoironHasan/bookservices/rest/route"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

// main is a main function
func main() {
	if errEnv := godotenv.Load(); errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	conf := config.New()

	db, errConn := persistence.NewDBConnection(conf.DBConfig)
	if errConn != nil {
		log.Fatalf("unable connect to database, %v", errConn)
	}

	repo := service.NewDBService(db)

	// grpc client
	clientConnBook, errClient := client.NewGRPCConn_Book(conf)
	if errClient != nil {
		log.Fatalf("grpc client unable connect to server, %v", errClient)
	}

	clientConnUser, errClient := client.NewGRPCConn_User(conf)
	if errClient != nil {
		log.Fatalf("grpc client unable connect to userservice, %v", errClient)
	}

	grpcClient := client.NewGRPCClient(clientConnBook, clientConnUser)

	command := cmd.NewCommand(
		cmd.WithConfig(conf),
		cmd.WithRepo(repo),
		cmd.WithGRPCClient(grpcClient),
	)

	app := cmd.NewCLI()
	app.Commands = command.Build()

	app.Action = func(ctx *cli.Context) error {
		router := route.NewRouter(
			route.WithConfig(conf),
			route.WithRepository(repo),
		).Init()

		shutdownTimeout := 10 * time.Second

		err := rest.RunHTTPServer(router, strconv.Itoa(conf.AppPort), shutdownTimeout)
		if err != nil {
			return err
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf("Unable to run CLI command, err: %v", err)
	}
}
