# Helldivers 2 dashboard

This is a dashboard for Helldivers 2, a game developed by Arrowhead Game Studios.


## Features

1. **Dashboards**
- [x] War room : A dashboard that shows the global campaign progress
- [x] Planet details : A dashboard that shows the details of a planet liberation progress

2. **Exporter**

Galaxy stats:

- `hde_galaxy_missions_won` : Number of missions won in the galaxy
- `hde_galaxy_missions_lost` : Number of missions lost in the galaxy
- `hde_galaxy_mission_time` : Cumulative mission time in the galaxy
- `hde_galaxy_bug_kills` : Number of bug kills in the galaxy
- `hde_galaxy_automaton_kills` : Number of automaton kills in the galaxy
- `hde_galaxy_illuminate_kills` : Number of illuminate kills in the galaxy
- `hde_galaxy_bullets_fired` : Number of bullets fired in the galaxy
- `hde_galaxy_bullets_hit` : Number of bullets hit in the galaxy
- `hde_galaxy_deaths` : Number of glorious helldivers sacrificed in the name of freedom !
- `hde_galaxy_revives` : Number of revive stratagems used
- `hde_galaxy_friendlies` : Remember, friendly fire isn't.
- `hde_galaxy_mission_success_rate` : Success rate of missions in the galaxy
- `hde_galaxy_accuracy` : Accuracy of helldivers in the galaxy

Planet stats:

- `hde_planet_health` : HP of a planet
- `hde_planet_max_health` : Max HP of a planet
- `hde_planet_players` : Number of players on a planet
- `hde_planet_regen_rate` : Regen rate of a planet
- `hde_planet_missions_won` : Number of missions won
- `hde_planet_missions_lost` : Number of missions lost
- `hde_planet_mission_time` : Cumulative mission time
- `hde_planet_bug_kills` : Number of bug kills
- `hde_planet_automaton_kills` : Number of automaton kills
- `hde_planet_illuminate_kills` : Number of illuminate kills
- `hde_planet_bullets_fired` : Number of bullets fired
- `hde_planet_bullets_hit` : Number of bullets hit
- `hde_planet_time_played` : Time played on a planet
- `hde_planet_deaths` : Number of glorious helldivers sacrificed in the name of freedom !
- `hde_planet_revives` : Number of revive stratagems used
- `hde_planet_friendlies` : Remember, friendly fire isn't.
- `hde_planet_mission_success_rate` : Success rate of missions
- `hde_planet_accuracy` : Accuracy of helldivers in the galaxy

# Installation

## Prerequisites

- `docker`
- `docker-compose` (can be skipped if your local docker version supports the `compose` command)

For Windows & Mac users, you can install Docker Desktop which includes both `docker` and `docker-compose`.
https://docs.docker.com/compose/install/#scenario-one-install-docker-desktop

## Steps

1. Clone the repository
2. Run `docker compose up --build` in the root directory of the repository


# Development

## Prerequisites

- `go` (1.22+)
- `earthly` (0.8.0+)

## Edit API Specifications

The API specifications are defined in the `specs.yaml` file. The API specifications are written in OpenAPI 3.0 format.
After editing the API specifications, run the following command to generate the API client:

```bash
earthly +codegen-client
```
The go client located at `pkg/client` will be updated to reflect the changes in the API specifications.


## Update planet JSON data

The planet JSON data is located at `data/planets.json`. To update the planet data, run the following command:

```bash
earthly +json-data
```
It will source planet data from the `helldivers-2/json` repository and update the `data/planets.json` file.

# Acknowledgements

- [Arrowhead Game Studios](https://arrowheadgamestudios.com/)
- [Helldivers community API](https://github.com/helldivers-2/api) for providing planet data and insights on the API structure.
