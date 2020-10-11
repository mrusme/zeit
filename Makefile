VERSION=0.0

all:
	go build -ldflags "-X github.com/mrusme/zeit/z.VERSION=$(VERSION)"
