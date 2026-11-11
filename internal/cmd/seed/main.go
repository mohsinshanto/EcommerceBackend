package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"path/filepath"
	"time"

	"ecommerce-backend/config"
	"ecommerce-backend/internal/module/product"
)

var (
	adjectives = []string{
		"Premium", "Smart", "Portable", "Elite", "Urban",
		"Wireless", "Pro", "Classic", "Modern", "Compact",
		"Essential", "Signature",
	}

	productNouns = map[string][]string{
		// Electronics
		"mobile":      {"Phone", "Handset", "Smartphone", "Device"},
		"laptop":      {"Laptop", "Notebook", "Ultrabook", "Workstation"},
		"audio":       {"Speaker", "Headphone", "Earbuds", "Soundbar"},
		"accessories": {"Mouse", "Keyboard", "Stand", "Charger", "Power Bank"},

		// Fashion
		"men's clothing":   {"Polo Shirt", "T-Shirt", "Denim Jacket", "Casual Shirt"},
		"women's clothing": {"Dress", "Blouse", "Kurti", "Top"},
		"shoes":            {"Sneakers", "Running Shoes", "Casual Shoes", "Formal Shoes"},
		"watches":          {"Wrist Watch", "Smart Watch", "Classic Watch"},
		"bags":             {"Backpack", "Handbag", "Shoulder Bag", "Travel Bag"},

		// Sports & Fitness
		"fitness equipment": {"Dumbbell", "Yoga Mat", "Resistance Band", "Exercise Bench"},
		"sportswear":        {"Sports T-Shirt", "Track Pants", "Training Shorts", "Sports Jacket"},
		"outdoor gear":      {"Camping Tent", "Hiking Backpack", "Sleeping Bag", "Water Bottle"},

		// Home & Living
		"furniture":          {"Chair", "Table", "Sofa", "Bookshelf"},
		"kitchen appliances": {"Blender", "Electric Kettle", "Rice Cooker", "Toaster"},
		"home decor":         {"Wall Clock", "Vase", "Decorative Mirror", "Photo Frame"},
		"lighting":           {"Table Lamp", "Ceiling Light", "Desk Lamp", "Floor Lamp"},
	}

	categoryDescriptions = map[string][]string{
		// Electronics
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

		// Fashion
		"men's clothing": {
			"Comfortable men's clothing designed for everyday wear and a modern look.",
			"Made with quality materials for casual outings and daily comfort.",
		},
		"women's clothing": {
			"Stylish women's clothing combining comfort and a contemporary design.",
			"Designed for everyday use with an elegant and comfortable fit.",
		},
		"shoes": {
			"Comfortable footwear designed for everyday activities and long-lasting use.",
			"A practical pair of shoes combining modern style and reliable comfort.",
		},
		"watches": {
			"A stylish timepiece with a clean design for everyday use.",
			"Combines practical functionality with a modern appearance.",
		},
		"bags": {
			"A practical bag with useful storage space for everyday essentials.",
			"Designed with a modern look and convenient storage compartments.",
		},

		// Sports & Fitness
		"fitness equipment": {
			"Designed to support home workouts and everyday fitness activities.",
			"A practical fitness accessory for building strength and maintaining an active lifestyle.",
		},
		"sportswear": {
			"Comfortable sportswear designed for training, exercise, and active lifestyles.",
			"Lightweight clothing that provides comfort and flexibility during workouts.",
		},
		"outdoor gear": {
			"Practical outdoor equipment designed for travel, hiking, and adventure.",
			"Built for outdoor activities with a convenient and durable design.",
		},

		// Home & Living
		"furniture": {
			"Functional furniture designed to improve comfort and everyday living spaces.",
			"A modern furniture piece combining practical use and attractive design.",
		},
		"kitchen appliances": {
			"Convenient kitchen equipment designed to make everyday cooking easier.",
			"A practical kitchen appliance for reliable everyday use.",
		},
		"home decor": {
			"An attractive home accessory designed to enhance your living space.",
			"Adds a decorative touch to modern homes and interiors.",
		},
		"lighting": {
			"Modern lighting designed to provide comfortable illumination for your space.",
			"A practical lighting solution with an attractive contemporary design.",
		},
	}

	// Existing Electronics images
	electronicsImages = []string{
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

type productCategory struct {
	name string
	id   uint
}

type seedGroup struct {
	name       string
	categories []productCategory
	images     []string
}

type seedConfig struct {
	products int
	batch    int
	category string
}

// Category IDs are the existing database IDs.
var electronicsCategories = []productCategory{
	{name: "laptop", id: 2},
	{name: "mobile", id: 6},
	{name: "audio", id: 7},
	{name: "accessories", id: 8},
}

var fashionCategories = []productCategory{
	{name: "men's clothing", id: 10},
	{name: "women's clothing", id: 11},
	{name: "shoes", id: 12},
	{name: "watches", id: 13},
	{name: "bags", id: 14},
}

var sportsCategories = []productCategory{
	{name: "fitness equipment", id: 16},
	{name: "sportswear", id: 17},
	{name: "outdoor gear", id: 18},
}

var homeCategories = []productCategory{
	{name: "furniture", id: 20},
	{name: "kitchen appliances", id: 21},
	{name: "home decor", id: 22},
	{name: "lighting", id: 23},
}

func buildImagePool(prefix string, count int) []string {
	images := make([]string, 0, count)

	for i := 1; i <= count; i++ {
		images = append(
			images,
			filepath.ToSlash(fmt.Sprintf("/products/%s%d.jpg", prefix, i)),
		)
	}

	return images
}

func getSeedGroups(category string) ([]seedGroup, error) {
	fashion := seedGroup{
		name:       "Fashion",
		categories: fashionCategories,
		images:     buildImagePool("fashion", 11),
	}

	sports := seedGroup{
		name:       "Sports & Fitness",
		categories: sportsCategories,
		images:     buildImagePool("fitness", 8),
	}

	home := seedGroup{
		name:       "Home & Living",
		categories: homeCategories,
		images:     buildImagePool("home", 10),
	}

	electronics := seedGroup{
		name:       "Electronics",
		categories: electronicsCategories,
		images:     electronicsImages,
	}

	switch category {
	case "new":
		return []seedGroup{fashion, sports, home}, nil
	case "fashion":
		return []seedGroup{fashion}, nil
	case "sports":
		return []seedGroup{sports}, nil
	case "home":
		return []seedGroup{home}, nil
	case "electronics":
		return []seedGroup{electronics}, nil
	case "all":
		return []seedGroup{electronics, fashion, sports, home}, nil
	default:
		return nil, fmt.Errorf("invalid category: %s", category)
	}
}

func main() {
	cfg := parseFlags()

	config.LoadEnv()
	config.ConnectDB()

	if err := config.DB.AutoMigrate(&product.Product{}); err != nil {
		log.Fatalf("failed to migrate product schema: %v", err)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	if cfg.products > 0 {
		if err := seedProducts(cfg, rng); err != nil {
			log.Fatalf("failed to seed products: %v", err)
		}
	}

	fmt.Println("Product seeding complete.")
}

func parseFlags() seedConfig {
	cfg := seedConfig{}

	flag.IntVar(&cfg.products, "products", 2000, "number of products per main category")
	flag.IntVar(&cfg.batch, "batch", 500, "insert batch size")
	flag.StringVar(&cfg.category, "category", "new", "category to seed: new, fashion, sports, home, electronics, all")

	flag.Parse()

	if cfg.batch <= 0 {
		cfg.batch = 500
	}

	return cfg
}

func seedProducts(cfg seedConfig, rng *rand.Rand) error {
	groups, err := getSeedGroups(cfg.category)
	if err != nil {
		return err
	}

	for _, group := range groups {
		if err := seedGroupProducts(group, cfg, rng); err != nil {
			return err
		}
	}

	return nil
}

func seedGroupProducts(group seedGroup, cfg seedConfig, rng *rand.Rand) error {
	fmt.Printf(
		"\nSeeding %d %s products in batches of %d...\n",
		cfg.products,
		group.name,
		cfg.batch,
	)

	for start := 0; start < cfg.products; start += cfg.batch {
		end := min(start+cfg.batch, cfg.products)

		products := make([]product.Product, 0, end-start)

		for i := start; i < end; i++ {
			category := group.categories[i%len(group.categories)]

			nouns := productNouns[category.name]
			descriptions := categoryDescriptions[category.name]

			adjective := adjectives[rng.Intn(len(adjectives))]
			noun := nouns[rng.Intn(len(nouns))]
			model := 100 + rng.Intn(900)

			categoryID := category.id

			products = append(products, product.Product{
				Name: fmt.Sprintf(
					"%s %s %d",
					adjective,
					noun,
					model,
				),
				Description: descriptions[rng.Intn(len(descriptions))],
				Price:       500 + float64(rng.Intn(14500)),
				Stock:       5 + rng.Intn(95),
				ImageURL:    group.images[rng.Intn(len(group.images))],

				Category:   category.name,
				CategoryID: &categoryID,
			})
		}

		if err := config.DB.CreateInBatches(products, cfg.batch).Error; err != nil {
			return err
		}

		fmt.Printf("Products seeded: %d/%d\n", end, cfg.products)
	}

	return nil
}

func init() {
	for i, path := range electronicsImages {
		electronicsImages[i] = filepath.ToSlash(path)
	}
}
