package bakap

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/cloud"
	"google.golang.org/cloud/storage"
)

// Service type
type Service struct {
	Interval   time.Duration
	StartTimes []time.Time
	Files      []File
	Bucket     string
	Account    string
	PrivateKey []byte
	PreScript  string
	PostScript string
	NamingFunc NamingFunc
	Async      bool

	bucket *storage.BucketHandle
}

// File type
type File struct {
	Src        string
	Dest       string
	PreScript  string
	PostScript string
}

// NamingFunc type
type NamingFunc func(File) string

// Run Bakap blocking service
func (srv *Service) Run() {
	var err error

	// fill default value
	if srv.NamingFunc == nil {
		srv.NamingFunc = generateName
	}

	googleConf := &jwt.Config{
		Email:      srv.Account,
		PrivateKey: srv.PrivateKey,
		TokenURL:   google.JWTTokenURL,
		Scopes:     []string{storage.ScopeReadWrite},
	}

	ctx := context.Background()

	client, err := storage.NewClient(ctx, cloud.WithTokenSource(googleConf.TokenSource(ctx)))
	if err != nil {
		log.Fatal(err)
	}
	srv.bucket = client.Bucket(srv.Bucket)

	srv.doBakap(srv.Files)
	if srv.Interval == 0 {
		return
	}
	for {
		time.Sleep(srv.Interval)
		srv.doBakap(srv.Files)
	}
}

func (srv *Service) doBakap(fs []File) {
	log.Println("bakap: start")
	for _, f := range fs {
		if srv.Async {
			go srv.bak(f)
		} else {
			srv.bak(f)
		}
	}
	log.Println("bakap: end")
}

func (srv *Service) bak(f File) error {
	log.Printf("bakap: kap %s => %s\n", f.Src, f.Dest)

	if f.PreScript != "" {
		runScript(f.PreScript)
	}

	fs, err := os.Open(f.Src)
	if err != nil {
		log.Printf("bakap: kap error %s\n", f.Src)
		return err
	}
	defer fs.Close()

	fileName := srv.NamingFunc(f)

	ctx := context.Background()
	w := srv.bucket.Object(fileName).NewWriter(ctx)
	if _, err := io.Copy(w, fs); err != nil {
		log.Printf("bakap: kap error %s\n", f.Src)
		return err
	}
	if err := w.Close(); err != nil {
		log.Printf("bakap: kap error %s\n", f.Src)
		return err
	}

	if f.PostScript != "" {
		runScript(f.PostScript)
	}

	log.Printf("bakap: kap end %s\n", f.Src)
	return nil
}

func generateName(f File) string {
	return fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), f.Dest)
}

func runScript(script string) error {
	return exec.Command("/bin/bash", "-c", script).Run()
}
