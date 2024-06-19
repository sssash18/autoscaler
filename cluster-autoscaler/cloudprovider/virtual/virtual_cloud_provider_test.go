package virtual

import "testing"

func TestReadClusterInfo(t *testing.T) {
	clusterInfoPath := "/tmp"
	readInitClusterInfo(clusterInfoPath)
}
