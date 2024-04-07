package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

  "github.com/Xide/helldivers2-dashboard/pkg/client"
)

var planetNames = map[int32]string{}

var flags *pflag.FlagSet = pflag.NewFlagSet("hde", pflag.ExitOnError)

func initLogger() {
  logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
  slog.SetDefault(logger)
}
func init() {
  viper.SetEnvPrefix("hde")
  viper.AutomaticEnv()

  viper.Set("collector_version", "0.0.1")
  flags.String("collector", "helldivers2-api", "Name of the collector")
  flags.String("api_url", "https://api.live.prod.thehelldiversgame.com/api", "URL of the API")
  flags.String("expose_address", ":9101", "Address to expose the metrics")
  flags.String("json_data_dir", "/data", "Directory where the static json data is stored")

  err := viper.BindPFlags(flags)
  if err != nil {
    panic(err)
  }
  initLogger()
}

var (
  planetHealth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_health",
    Help: "Health of the planet",
  }, []string{"planet"})

  planetMaxHealth = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_max_health",
    Help: "Max health of the planet",
  }, []string{"planet"})

  planetPlayers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_players",
    Help: "Number of players on the planet",
  }, []string{"planet"})

  planetRegenRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_regen_rate",
    Help: "Regen rate of the planet",
  }, []string{"planet"})

  apiRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
    Name: "hde_api_request_duration",
    Help: "Duration of the api request",
  }, []string{"route"})
  apiRequestStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
    Name: "hde_api_request_status",
    Help: "Status of the api request",
  }, []string{"route", "status"})

  galaxyMissionsWon = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_missions_won",
    Help: "Number of missions won in the galaxy",
  })

  galaxyMissionsLost = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_missions_lost",
    Help: "Number of missions lost in the galaxy",
  })

  galaxyMissionTime = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_mission_time",
    Help: "Time spent on missions in the galaxy",
  })

  galaxyBugKills = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_bug_kills",
    Help: "Number of bug kills in the galaxy",
  })

  galaxyAutomatonKills = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_automaton_kills",
    Help: "Number of automaton kills in the galaxy",
  })

  galaxyIlluminateKills = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_illuminate_kills",
    Help: "Number of illuminate kills in the galaxy",
  })

  galaxyBulletsFired = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_bullets_fired",
    Help: "Number of bullets fired in the galaxy",
  })

  galaxyBulletsHit = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_bullets_hit",
    Help: "Number of bullets hit in the galaxy",
  })

  galaxyTimePlayed = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_time_played",
    Help: "Time played in the galaxy",
  })

  galaxyDeaths = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_deaths",
    Help: "Number of deaths in the galaxy",
  })

  galaxyRevives = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_revives",
    Help: "Number of revives in the galaxy",
  })

  galaxyFriendlies = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_friendlies",
    Help: "Number of friendlies (fire?) in the galaxy",
  })

  galaxyMissionSuccessRate = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_mission_success_rate",
    Help: "Mission success rate in the galaxy",
  })

  galaxyAccuracy = prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "hde_galaxy_accuracy",
    Help: "Accuracy in the galaxy",
  })

  planetMissionsWon = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_missions_won",
    Help: "Number of missions won on the planet",
  }, []string{"planet"})
  planetMissionsLost = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_missions_lost",
    Help: "Number of missions lost on the planet",
  }, []string{"planet"})
  planetMissionTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_mission_time",
    Help: "Time spent on missions on the planet",
  }, []string{"planet"})
  planetBugKills = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_bug_kills",
    Help: "Number of bug kills on the planet",
  }, []string{"planet"})
  planetAutomatonKills = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_automaton_kills",
    Help: "Number of automaton kills on the planet",
  }, []string{"planet"})
  planetIlluminateKills = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_illuminate_kills",
    Help: "Number of illuminate kills on the planet",
  }, []string{"planet"})
  planetBulletsFired = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_bullets_fired",
    Help: "Number of bullets fired on the planet",
  }, []string{"planet"})
  planetBulletsHit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_bullets_hit",
    Help: "Number of bullets hit on the planet",
  }, []string{"planet"})
  planetTimePlayed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_time_played",
    Help: "Time played on the planet",
  }, []string{"planet"})
  planetDeaths = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_deaths",
    Help: "Number of deaths on the planet",
  }, []string{"planet"})
  planetRevives = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_revives",
    Help: "Number of revives on the planet",
  }, []string{"planet"})
  planetFriendlies = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_friendlies",
    Help: "Number of friendlies (fire?) on the planet",
  }, []string{"planet"})
  planetMissionSuccessRate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_mission_success_rate",
    Help: "Mission success rate on the planet",
  }, []string{"planet"})
  planetAccuracy = prometheus.NewGaugeVec(prometheus.GaugeOpts{
    Name: "hde_planet_accuracy",
    Help: "Accuracy on the planet",
  }, []string{"planet"})
)

// Fetch information from the 3 main endpoints:
// * Current war status (e.g. planet health, players, regen rate)
// * War info (e.g. max health of the planets)
// * War statistics (e.g. missions won, time played, etc.)
// Returns this information as 3 structs
// Fills prometheus histograms for HTTP queries
func fetch(cl client.ClientWithResponsesInterface) (*client.WarSeasonStatus, *client.WarSeasonInfo, *client.WarStatistics, error) {
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  tStart := time.Now()
  warStatus, err := cl.GetWarSeasonWarIdStatusWithResponse(ctx, 801)
  tEnd := time.Now()
  apiRequestDuration.WithLabelValues("war_status").Observe(tEnd.Sub(tStart).Seconds())
  if err != nil {
    slog.Error("Error fetching war status", slog.Any("error", err))
    return nil, nil, nil, err
  }
  slog.Info("Fetched war status", slog.Int("code", warStatus.StatusCode()), slog.Duration("duration", tEnd.Sub(tStart)))
  apiRequestStatus.WithLabelValues("war_status", fmt.Sprintf("%d", warStatus.StatusCode())).Inc()
  if warStatus.StatusCode() != 200 {
    slog.Error("Error code while fetching war status", slog.Int("code", warStatus.StatusCode()))
    return nil, nil, nil, fmt.Errorf("Error fetching war status")
  }
  warStatusRes := warStatus.JSON200

  tStart = time.Now()
  warInfo, err := cl.GetWarSeasonWarIdWarInfoWithResponse(ctx, 801)
  tEnd = time.Now()
  apiRequestDuration.WithLabelValues("war_info").Observe(tEnd.Sub(tStart).Seconds())
  if err != nil {
    slog.Error("Error fetching war info", slog.Any("error", err))
    return nil, nil, nil, err
  }
  slog.Info("Fetched war info", slog.Int("code", warInfo.StatusCode()), slog.Duration("duration", tEnd.Sub(tStart)))
  apiRequestStatus.WithLabelValues("war_info", fmt.Sprintf("%d", warInfo.StatusCode())).Inc()
  if warInfo.StatusCode() != 200 {
    slog.Error("Error code while fetching war info", slog.Int("code", warInfo.StatusCode()))
    return nil, nil, nil, fmt.Errorf("Error fetching war info")
  }
  warInfoRes := warInfo.JSON200

  tStart = time.Now()
  warStats, err := cl.GetStatsWarWarIdSummaryWithResponse(ctx, 801)
  tEnd = time.Now()
  apiRequestDuration.WithLabelValues("war_stats").Observe(tEnd.Sub(tStart).Seconds())
  if err != nil {
    slog.Error("Error fetching war stats", slog.Any("error", err))
    return nil, nil, nil, err
  }
  slog.Info("Fetched war stats", slog.Int("code", warStats.StatusCode()), slog.Duration("duration", tEnd.Sub(tStart)))
  apiRequestStatus.WithLabelValues("war_stats", fmt.Sprintf("%d", warStats.StatusCode())).Inc()
  if warStats.StatusCode() != 200 {
    slog.Error("Error code while fetching war stats", slog.Int("code", warStats.StatusCode()))
    return nil, nil, nil, fmt.Errorf("Error fetching war stats")
  }
  warStatsRes := warStats.JSON200

  return warStatusRes, warInfoRes, warStatsRes, nil
}

// Scrape the API and fill the prometheus metrics
// Returns an error if the API call fails
// called every 30 seconds
func scrape(cl client.ClientWithResponsesInterface) error {
  status, infos, stats, err := fetch(cl)
  if err != nil {
    return err
  }

  galaxyMissionsWon.Set(float64(stats.GalaxyStats.MissionsWon))
  galaxyMissionsLost.Set(float64(stats.GalaxyStats.MissionsLost))
  galaxyMissionTime.Set(float64(stats.GalaxyStats.MissionTime))
  galaxyBugKills.Set(float64(stats.GalaxyStats.BugKills))
  galaxyAutomatonKills.Set(float64(stats.GalaxyStats.AutomatonKills))
  galaxyIlluminateKills.Set(float64(stats.GalaxyStats.IlluminateKills))
  galaxyBulletsFired.Set(float64(stats.GalaxyStats.BulletsFired))
  galaxyBulletsHit.Set(float64(stats.GalaxyStats.BulletsHit))
  galaxyTimePlayed.Set(float64(stats.GalaxyStats.TimePlayed))
  galaxyDeaths.Set(float64(stats.GalaxyStats.Deaths))
  galaxyRevives.Set(float64(stats.GalaxyStats.Revives))
  galaxyFriendlies.Set(float64(stats.GalaxyStats.Friendlies))
  galaxyMissionSuccessRate.Set(float64(stats.GalaxyStats.MissionSuccessRate))
  galaxyAccuracy.Set(float64(stats.GalaxyStats.Accuracy))

  for _, planet := range stats.PlanetsStats {
    planetName, ok := planetNames[planet.PlanetIndex]
    if !ok {
      slog.Warn("Unknown planet", slog.Int("planet_id", int(planet.PlanetIndex)))
      continue
    }
    planetMissionsWon.WithLabelValues(planetName).Set(float64(planet.MissionsWon))
    planetMissionsLost.WithLabelValues(planetName).Set(float64(planet.MissionsLost))
    planetMissionTime.WithLabelValues(planetName).Set(float64(planet.MissionTime))
    planetBugKills.WithLabelValues(planetName).Set(float64(planet.BugKills))
    planetAutomatonKills.WithLabelValues(planetName).Set(float64(planet.AutomatonKills))
    planetIlluminateKills.WithLabelValues(planetName).Set(float64(planet.IlluminateKills))
    planetBulletsFired.WithLabelValues(planetName).Set(float64(planet.BulletsFired))
    planetBulletsHit.WithLabelValues(planetName).Set(float64(planet.BulletsHit))
    planetTimePlayed.WithLabelValues(planetName).Set(float64(planet.TimePlayed))
    planetDeaths.WithLabelValues(planetName).Set(float64(planet.Deaths))
    planetRevives.WithLabelValues(planetName).Set(float64(planet.Revives))
    planetFriendlies.WithLabelValues(planetName).Set(float64(planet.Friendlies))
    planetMissionSuccessRate.WithLabelValues(planetName).Set(float64(planet.MissionSuccessRate))
    planetAccuracy.WithLabelValues(planetName).Set(float64(planet.Accuracy))

  }
  for _, planet := range infos.PlanetInfos {
    planetName, ok := planetNames[planet.Index]
    if !ok {
      slog.Warn("Unknown planet", slog.Int("planet_id", int(planet.Index)))
      continue
    }
    planetMaxHealth.WithLabelValues(planetName).Set(float64(planet.MaxHealth))
  }
  for _, planet := range status.PlanetStatus {
    planetName, ok := planetNames[planet.Index]
    if !ok {
      slog.Warn("Unknown planet", slog.Int("planet_id", int(planet.Index)))
      continue
    }
    planetHealth.WithLabelValues(planetName).Set(float64(planet.Health))
    planetPlayers.WithLabelValues(planetName).Set(float64(planet.Players))
    planetRegenRate.WithLabelValues(planetName).Set(float64(planet.RegenPerSecond))
  }

  return nil
}

// Start the scraper, which will scrape the API every 30 seconds
// Started as a goroutine
func startScraper() {
  cl, err := client.NewClientWithResponses(viper.GetString("api_url"))
  if err != nil {
    panic(err)
  }
  slog.Info("Starting scraper")
  for {
    slog.Info("Performing scrape")
    err = scrape(cl)
    if err != nil {
      fmt.Println("Error scraping", err)
    }
    time.Sleep(30 * time.Second)
  }
}

// Load the static assets (planet names) from the json files
// Returns an error if the files cannot be loaded
func loadStaticAssets() error {
  slog.Info("Loading static assets")
  slog.Info("Loading planet names")
  planetsFile, err := os.Open(path.Join(viper.GetString("json_data_dir"), "planets.json"))
  if err != nil {
    slog.Error("Error loading planet names", slog.Any("error", err))
    return err
  }
  defer planetsFile.Close()
  planetData, err := io.ReadAll(planetsFile)
  if err != nil {
    slog.Error("Failed to read planet names", slog.Any("error", err))
    return err
  }
  err = json.Unmarshal(planetData, &planetNames)
  if err != nil {
    slog.Error("Failed to unmarshal planet names", slog.Any("error", err))
    return err
  }
  slog.Info("Loaded planet names")
  slog.Info("Loaded static assets")
  return nil
}

func main() {
  err := loadStaticAssets()
  if err != nil {
    panic(err)
  }
  	// Create a new registry.
	reg := prometheus.NewRegistry()

	// Register version collector.
	reg.MustRegister(version.NewCollector("hde"))

  reg.MustRegister(planetHealth)
  reg.MustRegister(planetMaxHealth)
  reg.MustRegister(planetRegenRate)
  reg.MustRegister(planetPlayers)
  reg.MustRegister(apiRequestDuration)
  reg.MustRegister(apiRequestStatus)
  reg.MustRegister(galaxyMissionsWon)
  reg.MustRegister(galaxyMissionsLost)
  reg.MustRegister(galaxyMissionTime)
  reg.MustRegister(galaxyBugKills)
  reg.MustRegister(galaxyAutomatonKills)
  reg.MustRegister(galaxyIlluminateKills)
  reg.MustRegister(galaxyBulletsFired)
  reg.MustRegister(galaxyBulletsHit)
  reg.MustRegister(galaxyTimePlayed)
  reg.MustRegister(galaxyDeaths)
  reg.MustRegister(galaxyRevives)
  reg.MustRegister(galaxyFriendlies)
  reg.MustRegister(galaxyMissionSuccessRate)
  reg.MustRegister(galaxyAccuracy)
  reg.MustRegister(planetMissionsWon)
  reg.MustRegister(planetMissionsLost)
  reg.MustRegister(planetMissionTime)
  reg.MustRegister(planetBugKills)
  reg.MustRegister(planetAutomatonKills)
  reg.MustRegister(planetIlluminateKills)
  reg.MustRegister(planetBulletsFired)
  reg.MustRegister(planetBulletsHit)
  reg.MustRegister(planetTimePlayed)
  reg.MustRegister(planetDeaths)
  reg.MustRegister(planetRevives)
  reg.MustRegister(planetFriendlies)
  reg.MustRegister(planetMissionSuccessRate)
  reg.MustRegister(planetAccuracy)

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{},
	))
  go startScraper()
  slog.Info("Starting server", slog.String("address", viper.GetString("expose_address")))
  err = http.ListenAndServe(viper.GetString("expose_address"), nil)
  if err != nil {
    slog.Error("Error starting server", slog.Any("error", err))
    panic(err)
  }
}
