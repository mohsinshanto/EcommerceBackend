package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"ecommerce-backend/config"
	"ecommerce-backend/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	firstNames   = []string{"Ayan", "Nabil", "Samiha", "Nusrat", "Rafi", "Mim", "Tahmid", "Raisa", "Shanto", "Sadia", "Farhan", "Tania"}
	lastNames    = []string{"Rahman", "Ahmed", "Karim", "Islam", "Hossain", "Sultana", "Chowdhury", "Hasan", "Akter", "Kabir"}
	adjectives   = []string{"Premium", "Smart", "Portable", "Elite", "Urban", "Wireless", "Pro", "Classic", "Modern", "Compact", "Essential", "Signature"}
	productNouns = map[string][]string{
		"mobile":      {"Phone", "Handset", "Smartphone", "Device"},
		"laptop":      {"Laptop", "Notebook", "Ultrabook", "Workstation"},
		"audio":       {"Speaker", "Headphone", "Earbuds", "Soundbar"},
		"accessories": {"Mouse", "Keyboard", "Stand", "Charger", "Power Bank"},
	}
	categoryDescriptions = map[string][]string{
		"mobile": {
			"Built for smooth everyday performance with a bright display and dependable battery life.",
			"A balanced phone for calls, streaming, browsing, and daily work without hassle.",
		},
		"laptop": {
			"Designed for study, work, and entertainment with reliable performance throughout the day.",
			"A practical laptop with comfortable multitasking and a clean modern design.",
		},
		"audio": {
			"Clear sound output with a lightweight design that fits home, office, and travel use.",
			"An easy everyday audio choice for music, movies, and casual listening sessions.",
		},
		"accessories": {
			"A useful accessory that improves comfort, setup flexibility, and day-to-day convenience.",
			"Built for regular use with a simple design and dependable everyday function.",
		},
	}
	imagePool = []string{
		"/products/acer.avif",
		"/products/asusrog.jpeg",
		"/products/beats.jpeg",
		"/products/bluetooth.jpg",
		"/products/dell.jpg",
		"/products/electric.jpg",
		"/products/gamingmouse.jpeg",
		"/products/hp.jpeg",
		"/products/iphone.jpg",
		"/products/jbl.jpeg",
		"/products/lenovo.jpeg",
		"/products/logitech.jpg",
		"/products/macbook.jpg",
		"/products/macbookair.jpeg",
		"/products/oneplus.jpeg",
		"/products/powerbank.avif",
		"/products/redmi.jpeg",
		"/products/samsung.jpg",
		"/products/slimkeyboard.jpg",
		"/products/smart.jpg",
		"/products/sony.jpg",
		"/products/speaker.jpeg",
		"/products/tabletpro.jpg",
		"/products/wireless-headphones.avif",
		"/products/wooden.jpg",
	}
)

type seedConfig struct {
	users    int
	products int
	batch    int
	prefix   string
	password string
}

func main() {
	cfg := parseFlags()

	config.LoadEnv()
	config.LoadJWTSecret()
	config.ConnectDB()

	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Cart{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		log.Fatalf("failed to migrate schema: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash seed password: %v", err)
	}

	if cfg.users > 0 {
		if err := seedUsers(cfg, rng, string(hashedPassword)); err != nil {
			log.Fatalf("failed to seed users: %v", err)
		}
	}

	if cfg.products > 0 {
		if err := seedProducts(cfg, rng); err != nil {
			log.Fatalf("failed to seed products: %v", err)
		}
	}

	fmt.Println("Seeding complete.")
	fmt.Printf("Seeded user password: %s\n", cfg.password)
	fmt.Printf("Seed prefix used for users: %s\n", cfg.prefix)
}

func parseFlags() seedConfig {
	defaultPrefix := fmt.Sprintf("seed%d", time.Now().Unix())

	cfg := seedConfig{}
	flag.IntVar(&cfg.users, "users", 10000, "number of users to seed")
	flag.IntVar(&cfg.products, "products", 10000, "number of products to seed")
	flag.IntVar(&cfg.batch, "batch", 500, "insert batch size")
	flag.StringVar(&cfg.prefix, "prefix", defaultPrefix, "email prefix to keep seeded users unique")
	flag.StringVar(&cfg.password, "password", "SeedUser123!", "shared password for all seeded users")
	flag.Parse()

	if cfg.batch <= 0 {
		cfg.batch = 500
	}

	return cfg
}

func seedUsers(cfg seedConfig, rng *rand.Rand, hashedPassword string) error {
	fmt.Printf("Seeding %d users in batches of %d...\n", cfg.users, cfg.batch)

	for start := 0; start < cfg.users; start += cfg.batch {
		end := min(start+cfg.batch, cfg.users)
		users := make([]models.User, 0, end-start)

		for i := start; i < end; i++ {
			firstName := firstNames[rng.Intn(len(firstNames))]
			lastName := lastNames[rng.Intn(len(lastNames))]
			users = append(users, models.User{
				Name:     firstName + " " + lastName,
				Email:    fmt.Sprintf("%s_user_%06d@example.com", cfg.prefix, i+1),
				Password: hashedPassword,
				IsAdmin:  false,
			})
		}

		if err := config.DB.CreateInBatches(users, cfg.batch).Error; err != nil {
			return err
		}

		fmt.Printf("  Users seeded: %d/%d\n", end, cfg.users)
	}

	return nil
}

func seedProducts(cfg seedConfig, rng *rand.Rand) error {
	fmt.Printf("Seeding %d products in batches of %d...\n", cfg.products, cfg.batch)

	categories := orderedCategories()

	for start := 0; start < cfg.products; start += cfg.batch {
		end := min(start+cfg.batch, cfg.products)
		products := make([]models.Product, 0, end-start)

		for i := start; i < end; i++ {
			category := categories[i%len(categories)]
			nouns := productNouns[category]
			descriptions := categoryDescriptions[category]
			adjective := adjectives[rng.Intn(len(adjectives))]
			noun := nouns[rng.Intn(len(nouns))]
			model := 100 + rng.Intn(900)

			products = append(products, models.Product{
				Name:        fmt.Sprintf("%s %s %d", adjective, noun, model),
				Description: descriptions[rng.Intn(len(descriptions))],
				Price:       500 + float64(rng.Intn(14500)),
				Stock:       5 + rng.Intn(95),
				ImageURL:    imagePool[rng.Intn(len(imagePool))],
				Category:    category,
			})
		}

		if err := config.DB.CreateInBatches(products, cfg.batch).Error; err != nil {
			return err
		}

		fmt.Printf("  Products seeded: %d/%d\n", end, cfg.products)
	}

	return nil
}

func orderedCategories() []string {
	categories := make([]string, 0, len(productNouns))
	for category := range productNouns {
		categories = append(categories, category)
	}

	// Keep the output stable across runs.
	for i := 0; i < len(categories); i++ {
		for j := i + 1; j < len(categories); j++ {
			if strings.Compare(categories[i], categories[j]) > 0 {
				categories[i], categories[j] = categories[j], categories[i]
			}
		}
	}

	return categories
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// Sanity check the image paths we reuse during seeding.
	for i, path := range imagePool {
		imagePool[i] = filepath.ToSlash(path)
	}
}
