package main

import (
	"context"
	"drones/internal/pkg/app"
	"log"
)

// @title Заявки контроля маршрутов БПЛА
// @version 0.0-0

// @host 127.0.0.1:8000
// @schemes http
// @BasePath /

func main() {
	log.Println("Application start!")

	a, err := app.New(context.Background())
	if err != nil {
		log.Println(err)

		return
	}

	a.StartServer()

	log.Println("Application terminated!")

	// endpoint := "127.0.0.1:9000"
	// accessKeyID := "minioadmin"
	// secretAccessKey := "minioadmin"
	// useSSL := false

	// minioClient, err := minio.New(endpoint, &minio.Options{
	// 	Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
	// 	Secure: useSSL,
	// })

	// if err != nil {
	// 	log.Fatalln(err)

	// 	return
	// }

	// log.Printf("%#v\n", minioClient)

	// object, err := minioClient.GetObject(context.Background(), "regionimages", "kuzminki.jpg", minio.GetObjectOptions{})
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer object.Close()

	// localFile, err := os.Create("local-file.jpg")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer localFile.Close()

	// if _, err = io.Copy(localFile, object); err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

}
