# FileCache - Efficient Go File Caching

FileCache is a Go module that provides a simple file caching mechanism. It allows you to efficiently manage and retrieve files, with the option to cache them in memory or use a streaming approach for large files.

## Features

- In-memory caching of small files.
- Streaming for large files to avoid excessive memory usage.
- Efficient file updates and deletions.

## Installation

```bash
go get -u github.com/NIR3X/filecache
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/NIR3X/filecache"
)

func main() {
	// Create a new FileCache instance with a specified maximum cache size
	f := filecache.NewFileCache(2 * 1024 * 1024)

	// Update the cache with a file
	err := f.Update("/path/to/your/file.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Get a reader for the file from the cache
	reader, writer, err := f.Get("/path/to/your/file.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Use the reader to read the file content

	// Remember to close the writer if it's not nil

	// ...

	// Delete a file from the cache
	f.Delete("/path/to/your/file.txt")
}
```

## License

[![GNU AGPLv3 Image](https://www.gnu.org/graphics/agplv3-155x51.png)](https://www.gnu.org/licenses/agpl-3.0.html)

This program is Free Software: You can use, study share and improve it at your
will. Specifically you can redistribute and/or modify it under the terms of the
[GNU Affero General Public License](https://www.gnu.org/licenses/agpl-3.0.html) as
published by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
