### .env
Muss immer der Pfad angegeben werden, von dem aus man startet. Wenn man mit dem Command "go run ./cmd/api", also aus dem Root startet, ist der Pfad relativ gesehen, also er liegt im selben directory, heißt in main. go --> godotenv.Load(".env")

Wenn man aber in das directory direct hineingeht, muss der Pfad so sein:
--> godotenv.Load("../../.env") 