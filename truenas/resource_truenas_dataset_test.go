package truenas

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_datasetPathString(t *testing.T) {
	testcases := []struct {
		path     datasetPath
		expected string
	}{
		{ path: datasetPath{ Pool: "Tank", Parent: "home", Name: "Test" }, expected: "Tank/home/Test" },
		{ path: datasetPath{ Pool: "Tank", Parent: "", Name: "Test" }, expected: "Tank/Test" },
		{ path: datasetPath{ Pool: "Tank", Parent: "//home/sub//", Name: "Test" }, expected: "Tank/home/sub/Test" },
		{ path: datasetPath{ Pool: "TankV2", Parent: "/home/", Name: "Test" }, expected: "TankV2/home/Test" },
	}

	for _, c := range testcases {
		actual := c.path.String()
		assert.Equal(t, actual, c.expected)
	}
}

func Test_newDatasetPath(t *testing.T) {
	testcases := []struct {
		path       string
		expected datasetPath
	}{
		{ expected: datasetPath{ Pool: "Tank", Parent: "home", Name: "Test" }, path: "Tank/home/Test" },
		{ expected: datasetPath{ Pool: "Tank", Parent: "", Name: "Test" }, path: "Tank/Test" },
		{ expected: datasetPath{ Pool: "Tank", Parent: "home/sub", Name: "Test" }, path: "Tank/home/sub/Test" },
		{ expected: datasetPath{ Pool: "TankV2", Parent: "home", Name: "Test" }, path: "TankV2/home/Test" },
	}

	for _, c := range testcases {
		actual := newDatasetPath(c.path)
		assert.Equal(t, actual, c.expected)
	}
}