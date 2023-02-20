# Migration

Go module for Estafette Builds and Releases migration from one SCM source to other

## API

API specification for migration is defined in [migration.openapi.yaml](https://github.com/estafette/estafette-ci-api/blob/main/migration.openapi.yaml)

## Client

Follow below steps to use go client

- Install [migration client](https://github.com/estafette/migration)

  ```shell
  go get github.com/estafette/migration
  ```

- Create clientID and secret in Estafette UI
- Use client to queue migration task

  Example:

  ```go
  package main

  import "github.com/estafette/migration"

  func main()  {
      client := migration.NewClient("https://api.estafette.io", "<Client-ID>", "<Client-Secret>")
      callbackURL := "https://your-callback-url"
      req := migration.TaskRequest{
          ID:          "existing-id", // optional if creating a new migration task
          FromSource:  "bitbucket.com",
          FromOwner:   "owner1",
          FromName:    "repo1",
          ToSource:    "github.com",
          ToOwner:     "owner1",
          ToName:      "repo1",
          CallbackURL: &callbackURL, // optional
          Restart:     migration.BuildLogsStage, // optional, default is LastStage (handled by the server)
      }
      task, err := client.QueueMigration(req)
      if err != nil {
          panic(err)
      }
      fmt.Sprintf("Migration task %v queued", task.ID)
  }
  ```

- Use client to get status of migration task

  Example:

  ```go
  package main

  import "github.com/estafette/migration"

  func main()  {
      client := migration.NewClient("https://api.estafette.io", "<Client-ID>", "<Client-Secret>")
      callbackURL := "https://your-callback-url"
      req := migration.TaskRequest{
          ID:          "existing-id", // optional if creating a new migration task
          FromSource:  "bitbucket.com",
          FromOwner:   "owner1",
          FromName:    "repo1",
          ToSource:    "github.com",
          ToOwner:     "owner1",
          ToName:      "repo1",
          CallbackURL: &callbackURL, // optional
          Restart:     migration.BuildLogsStage, // optional, default is LastStage (handled by the server)
      }
      task, err := client.QueueMigration(req)
      if err != nil {
          panic(err)
      }
      fmt.Sprintf("Migration task %v queued", task.ID)
  }
  ```
