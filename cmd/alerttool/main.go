package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	yaml "gopkg.in/yaml.v3"

	"github.com/grafana/loki/pkg/logql"
	"github.com/grafana/loki/pkg/ruler/manager"
	"github.com/prometheus/prometheus/pkg/rulefmt"
)

type Config struct {
	inFile, outFile string
}

func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.inFile, "if", "", "file to read")
	f.StringVar(&c.outFile, "of", "", "file to write")
}

func (c *Config) Validate() error {
	if c.inFile == "" {
		return fmt.Errorf("in-file (if) must not be empty")
	}
	if c.outFile == "" {
		return fmt.Errorf("out-file (of) must not be empty")
	}
	return nil
}

func main() {
	var config Config
	config.RegisterFlags(flag.CommandLine)
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}

	var loader manager.GroupLoader

	groups, errs := loader.Load(config.inFile)
	if len(errs) > 0 {
		log.Fatal(errs[0])
	}

	for i := range groups.Groups {
		grp := &groups.Groups[i]
		var updated []rulefmt.RuleNode
		for j := range grp.Rules {
			rule := &grp.Rules[j]
			// TODO(owen-d): handle recording rules gracefully -- ignore them.

			parsed, err := logql.ParseExpr(rule.Expr.Value)
			if err != nil {
				log.Fatal(
					fmt.Errorf(
						"error parsing (grp: %s, rule: %s): %s",
						grp.Name,
						rule.Alert.Value,
						err.Error(),
					),
				)
			}
			sampleExpr, ok := parsed.(logql.SampleExpr)
			if !ok {
				log.Fatal(fmt.Errorf(
					"expected a sampleExpr: (grp: %s, rule: %s)",
					grp.Name,
					rule.Alert.Value,
				))
			}
			dms, err := logql.MetaAlert(sampleExpr)
			if err != nil {
				log.Fatal(fmt.Errorf(
					"error creating dead mans switch (grp: %s, rule: %s): %s",
					grp.Name,
					rule.Alert.Value,
					err.Error(),
				))
			}
			if dms == nil {
				continue
			}
			rule.Expr.SetString(dms.String())
			rule.Alert.SetString(rule.Alert.Value + "_failsafe")
			updated = append(updated, *rule)
		}

		grp.Rules = updated
	}

	out, err := yaml.Marshal(groups)
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(config.outFile, out, 0644); err != nil {
		log.Fatal(err)
	}
}
