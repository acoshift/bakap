package bakap

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

// Config type
type Config struct {
	Interval   time.Duration
	Files      []File
	Bucket     string
	Account    string
	PrivateKey []byte
}

// File type
type File struct {
	Src  string
	Dest string
}

var bucket *storage.BucketHandle

// Start Bakap
func Start(c Config) {
	var err error

	googleConf := &jwt.Config{
		Email:      c.Account,
		PrivateKey: c.PrivateKey,
		TokenURL:   google.JWTTokenURL,
		Scopes:     []string{storage.ScopeReadWrite},
	}

	ctx := context.Background()

	client, err := storage.NewClient(ctx, cloud.WithTokenSource(googleConf.TokenSource(ctx)))
	if err != nil {
		log.Fatal(err)
	}
	bucket = client.Bucket(c.Bucket)

	doBakap(c.Files)
	if c.Interval == 0 {
		return
	}
	for {
		time.Sleep(c.Interval)
		doBakap(c.Files)
	}
}

func doBakap(fs []File) {
	for i, f := range fs {
		log.Println("bakap: start ", time.Now())
		log.Printf("bakap: kap %d, %s => %s\n", i, f.Src, f.Dest)
		uploadFile(f)
		log.Println("bakap: end ", time.Now())
	}
}

func uploadFile(f File) error {
	fs, err := os.Open(f.Src)
	if err != nil {
		return err
	}
	defer fs.Close()

	fileName := fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), f.Dest)

	ctx := context.Background()
	w := bucket.Object(fileName).NewWriter(ctx)
	if _, err := io.Copy(w, fs); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return nil
}
