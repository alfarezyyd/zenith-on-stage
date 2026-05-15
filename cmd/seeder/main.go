package main

import (
	"go-web-boilerplate/cmd/injector"
	"log"
	"strings"
)

func main() {
	viperConfig := injector.NewViperConfig()
	databaseCredential := injector.NewDatabaseCredentials(viperConfig)
	gormInstance := injector.NewDatabaseConnection(databaseCredential)
	viper.AutomaticEnv()

	// Read SEEDER env
	seederEnv := viper.GetString("SEEDER")

	if seederEnv == "" {
		log.Fatal("Please set SEEDER environment variable, e.g. SEEDER=all or SEEDER=UserSeeder")
	}

	seederNames := []string{}
	if strings.ToLower(seederEnv) == "all" {
		for name := range seeders.SeederRegistry {
			seederNames = append(seederNames, name)
		}
	} else {
		seederNames = append(seederNames, seederEnv)
	}

	for _, name := range seederNames {
		s, err := seeders.GetSeeder(name)
		if err != nil {
			log.Println(err)
			continue
		}
		logrus.Debug("Running seeder:", name)
		if err := s.Run(gormInstance); err != nil {
			log.Println("Seeder failed:", err)
		} else {
			logrus.Debug("Seeder completed:", name)
		}
	}
}

func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	sqlDB.Close()
}
