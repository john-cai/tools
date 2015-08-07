# McRedis

Mostly just an abstraction and isolation of the [already existing](https://github.com/sendgrid/catalyst/tree/master/mcredis) mcredis library.

### Usage
``` go
import (
	"fmt"
	"time"

	"github.com/sendgrid/mcredis"

	"github.com/garyburd/redigo/redis"
)

func main() {
	config := mcredis.RedisConfig{
		Nodes:       []string{"192.16.256.0:26379", "192.16.256.1:26379", "192.16.256.2:26379"},
		MasterName:  "my-cluster-name",
		Timeout:     time.Duration(30) * time.Second,
		CallTimeout: time.Duration(1) * time.Second,
		PoolSize:    5,
		MaxIdle:     3,
	}

	var r mcredis.RedisInstance // for type demonstration purposes, also thread and goroutine safe
	r, err = mcredis.NewMcRedis(config)
	if err != nil {
		fmt.Println("Sad panda")
	}

	err = r.Set("key", "value")
	if err != nil {
		fmt.Println("You are a terrible programmer")
	}

	value, err := r.Get("key")
	if err != nil {
		fmt.Println("This is just getting embarrassing")
	}
	stringValue, err := redis.String(value, nil)
	fmt.Println(stringValue)
}
```

### Mocking

``` go
import (
	"fmt"
	"time"

	"github.com/sendgrid/mcredis"
	"github.com/sendgrid/mcredis/fake"

	"github.com/garyburd/redigo/redis"
)

func main() {
	r := mcredis.RedisInstance{fake.NewFakeCommander()} 
	if err != nil {
		fmt.Println("Sad panda")
	}

	err = r.Set("key", "value")
	if err != nil {
		fmt.Println("You are a terrible programmer")
	}

	value, err := r.Get("key")
	if err != nil {
		fmt.Println("This is just getting embarrassing")
	}
	stringValue, err := redis.String(value, nil)
	fmt.Println(stringValue)
}
```