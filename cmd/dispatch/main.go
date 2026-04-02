package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-dispatch/internal/server";"github.com/stockyard-dev/stockyard-dispatch/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="8560"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./dispatch-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("dispatch: %v",err)};defer db.Close();srv:=server.New(db)
fmt.Printf("\n  Dispatch — task dispatch and routing\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("dispatch: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
