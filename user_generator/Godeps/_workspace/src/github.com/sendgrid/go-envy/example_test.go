package envy

// "examples" are added here to improve the output messages
type ExampleConfig struct {
	Host       string   `example:"localhost"`
	Port       int      `example:"9000"`
	Debug      bool     `example:"false"`
	StringList []string `example:"string1,string2,str3"`
	IntList    []int    `example:"1,-1,32313123"`

	INeedThisNameToBeLongBecauseItMakesSenseForMe string `name:"TESTING"`
}

func ExampleConfigWithPrefix() {
	// Env Variables Sample:
	//
	// KAMTA_HOST=localhost KAMTA_PORT=9000 KAMTA_DEBUG=true TESTING=something

	config := ExampleConfig{}
	LoadWithPrefix("KAMTA_", &config)
}
