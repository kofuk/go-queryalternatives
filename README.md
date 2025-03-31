# query-alternatives

query-alternatives is a Go package that provides a simple parser for `update-alternatives --query`.

## Usage

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/kofuk/go-queryalternatives"
)

func main() {
	queryOutput := `Name: java
Link: /usr/bin/java
Slaves:
 java.1.gz /usr/share/man/man1/java.1.gz
Status: auto
Best: /usr/lib/jvm/java-21-openjdk-amd64/bin/java
Value: /usr/lib/jvm/java-21-openjdk-amd64/bin/java

Alternative: /usr/lib/jvm/java-21-openjdk-amd64/bin/java
Priority: 2111
Slaves:
 java.1.gz /usr/lib/jvm/java-21-openjdk-amd64/man/man1/java.1.gz
 java.ja.1.gz /usr/lib/jvm/java-21-openjdk-amd64/man/ja/man1/java.1.gz

Alternative: /usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java
Priority: 1081
Slaves:
 java.1.gz /usr/lib/jvm/java-8-openjdk-amd64/jre/man/man1/java.1.gz
`

	alternatives, err := queryalternatives.ParseString(queryOutput)
	if err != nil {
		log.Fatal(err)
	}

	resultJson, _ := json.MarshalIndent(alternatives, "", "  ")
	fmt.Println(string(resultJson))
	// Output:
	// {
	//  "Name": "java",
	//  "Link": "/usr/bin/java",
	//  "Slaves": {
	//    "java.1.gz": "/usr/share/man/man1/java.1.gz"
	//  },
	//  "Status": "auto",
	//  "Best": "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
	//  "Value": "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
	//  "Alternatives": [
	//    {
	//      "Path": "/usr/lib/jvm/java-21-openjdk-amd64/bin/java",
	//      "Priority": 2111,
	//      "Slaves": {
	//        "java.1.gz": "/usr/lib/jvm/java-21-openjdk-amd64/man/man1/java.1.gz",
	//        "java.ja.1.gz": "/usr/lib/jvm/java-21-openjdk-amd64/man/ja/man1/java.1.gz"
	//      }
	//    },
	//    {
	//      "Path": "/usr/lib/jvm/java-8-openjdk-amd64/jre/bin/java",
	//      "Priority": 1081,
	//      "Slaves": {
	//        "java.1.gz": "/usr/lib/jvm/java-8-openjdk-amd64/jre/man/man1/java.1.gz"
	//      }
	//    }
	//  ]
	//}
}
```

If you want this library to query the alternatives using `update-alternatives` command, you can use the `queryalternatives.Query` function.

## License

MIT
