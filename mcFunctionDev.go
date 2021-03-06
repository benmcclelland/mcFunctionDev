package main

import (
	//"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/BurntSushi/toml"
	mcshapes "github.com/GreenSeaTurtle/mcFunctionDev/mcShapes"
)

// mcFunctionPath struct for reading various things from the init file
// Note that fields must start with a capital letter!!!!!!
type mcFunctionPath struct {
	Title          string
	MCSavesDir     string `toml:"mc_saves_dir"`
	MCFunctionsDir string `toml:"mc_world_functions_dir"`
}

func main() {
	// Read and extract information from the init file.  Right now, the
	// only information in the init file is the path to the Minecraft
	// functions directory on this system.  The waterfall files are
	// written directly to the game directory which saves time and hassle
	// of copying files.  The path is split into two strings just because
	// it is typically a long path.
	//
	// The TOML package is used to read and parse the init file.
	//    github.com/BurntSushi/toml
	//
	gopath := os.Getenv("GOPATH")
	infile := gopath + "/mc_function_dev.init"
	//fmt.Println("infile = " + infile)
	var config mcFunctionPath
	if _, err := toml.DecodeFile(infile, &config); err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Printf("Title: %s\n", config.Title)
	//fmt.Printf("mc_saves_dir: %s\n", config.MCSavesDir)
	//fmt.Println("mc_world_functions_dir = " + config.MCFunctionsDir)

	// Keep this for now as an example of how to get and process
	// execution line arguments.
	//flag.StringVar(&mcSavesDir, "s", "~", "Minecraft saves directory")
	//flag.StringVar(&mcWorldFuncDir, "w", "mc", "Minecraft functions directory")
	//flag.Parse()

	basepath := path.Join(config.MCSavesDir, config.MCFunctionsDir)
	//fmt.Println("basepath = " + basepath)
	err := BuildWaterFalls(basepath)
	if err != nil {
		log.Fatalln(err)
	}

	err = BuildLavaFalls(basepath)
	if err != nil {
		log.Fatalln(err)
	}

	err = BuildRollerCoasterFalls(basepath)
	if err != nil {
		log.Fatalln(err)
	}

	err = rmFalls(basepath)
	if err != nil {
		log.Fatalln(err)
	}

	err = ClearForWall(basepath)
	if err != nil {
		log.Fatalln(err)
	}

	err = CreateSphere(basepath, 20, "glass", "lava")
	if err != nil {
		log.Fatalln(err)
	}
	err = CreateSphere(basepath, 20, "sea_lantern", "nothing")
	if err != nil {
		log.Fatalln(err)
	}

}

//BuildWaterFalls builds n, s, e, w waterfalls
func BuildWaterFalls(basepath string) error {
	origin := mcshapes.XYZ{X: 0, Y: 0, Z: -2}

	for _, direction := range []string{"north", "east", "south", "west"} {
		// Minecraft functions must have a suffix of ".mcfunction"
		fname := path.Join(basepath, "waterfall_"+direction) + ".mcfunction"
		f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("open %v: %v", fname, err)
		}
		defer f.Close()

		obj := mcshapes.NewMCObject(mcshapes.WithOrientation(direction))
		wf := CreateWaterfall(origin, obj)
		err = mcshapes.WriteShapes(f, wf)
		if err != nil {
			return fmt.Errorf("build waterfall: %v", err)
		}
	}

	return nil
}

//BuildLavaFalls builds n, s, e, w lava falls
func BuildLavaFalls(basepath string) error {
	origin := mcshapes.XYZ{X: 0, Y: 0, Z: -2}

	for _, direction := range []string{"north", "east", "south", "west"} {
		// Minecraft functions must have a suffix of ".mcfunction"
		fname := path.Join(basepath, "lavafall_"+direction) + ".mcfunction"
		f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("open %v: %v", fname, err)
		}
		defer f.Close()

		obj := mcshapes.NewMCObject(
			mcshapes.WithOrientation(direction),
			mcshapes.WithType("lavafall"))
		wf := CreateWaterfall(origin, obj)
		err = mcshapes.WriteShapes(f, wf)
		if err != nil {
			return fmt.Errorf("build lavafall: %v", err)
		}
	}

	return nil
}

//BuildRollerCoasterFalls builds two falls next to each other separated by only one
// block. It adds redstone and track to make it a roller coaster ride.
func BuildRollerCoasterFalls(basepath string) error {
	// Create the file that will contain both the north and south waterfalls.
	fname := path.Join(basepath, "waterfall_rc_north_south.mcfunction")
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open %v: %v", fname, err)
	}
	defer f.Close()

	// Build the north fall - faces south, runs west to east.
	origin := mcshapes.XYZ{X: 2, Y: 0, Z: -2}
	direction := "north"
	obj := mcshapes.NewMCObject(mcshapes.WithOrientation(direction))
	wf := CreateWaterfall(origin, obj)
	err = mcshapes.WriteShapes(f, wf)
	if err != nil {
		return fmt.Errorf("build waterfall rc north fall: %v", err)
	}

	// Build the south fall - faces north, runs west to east.
	// The north and south falls are -1 blocks apart and so they share the same blocks for
	// the front of the basin. In fact, the south falls overwrites what the north falls
	// for the front of the basin.
	// This ends up producing two sheets of water, one block apart. The roller coaster
	// goes between those sheets.
	origin = mcshapes.XYZ{X: 2, Y: 0, Z: 2}
	direction = "south"
	obj = mcshapes.NewMCObject(mcshapes.WithOrientation(direction))
	wf = CreateWaterfall(origin, obj)
	err = mcshapes.WriteShapes(f, wf)
	if err != nil {
		return fmt.Errorf("build waterfall rc south fall: %v", err)
	}

	// At this point we have two waterfalls facing each other, separated by one row of blocks,
	// i.e. the front of the basin which defaults to sandstone.
	// Change those blocks to be redstone in preparation for putting tracks on them.
	// Replace the sandstone with redstone to power the rails.
	width := 102
	corner1 := mcshapes.XYZ{X: origin.X, Y: origin.Y, Z: origin.Z - 4}
	corner2 := mcshapes.XYZ{X: origin.X + width - 1, Y: origin.Y, Z: origin.Z - 4}
	b := mcshapes.NewBox(mcshapes.WithCorner1(corner1), mcshapes.WithCorner2(corner2),
		mcshapes.WithSurface("minecraft:redstone_block"))
	err = b.WriteShape(f)
	if err != nil {
		return fmt.Errorf("build waterfall rc redstone: %v", err)
	}

	// Lay down powered track on top of the redstone.
	corner1.Y += 1
	corner2.Y += 1
	b = mcshapes.NewBox(mcshapes.WithCorner1(corner1), mcshapes.WithCorner2(corner2),
		mcshapes.WithSurface("minecraft:golden_rail"))
	err = b.WriteShape(f)
	if err != nil {
		return fmt.Errorf("build waterfall rc track: %v", err)
	}

	return nil
}

// rmFalls removes north, south, east, west waterfalls
// After placing a water or lava fall somewhere in the Minecraft game, it is sometimes
// necessary to remove it. Perhaps, for example, it was placed in the wrong location and
// needs to be moved. This function writes out a Minecraft function to do this using the
// Minecraft fill command.
// An example of this command is:
//    fill ~0 ~0 ~-2 ~101 ~30 ~-6 minecraft:air
// The ~ refers to the players current position in the game.
// Yes, a fall could be removed by hand inside the game, but this is very tedious, thus
// the need for this function.
func rmFalls(basepath string) error {
	origin := mcshapes.XYZ{X: 0, Y: 0, Z: -2}

	for _, direction := range []string{"north", "east", "south", "west"} {
		// Minecraft functions must have a suffix of ".mcfunction"
		fname := path.Join(basepath, "rm_fall_"+direction) + ".mcfunction"
		f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("rm_fall open %v: %v", fname, err)
		}
		defer f.Close()

		width := 100
		height := 30
		// Use a loop because Minecraft has a limit on total number of blocks per fill command.
		for h := 0; h <= height; h++ {
			corner1 := mcshapes.XYZ{X: origin.X, Y: origin.Y + h, Z: origin.Z}
			corner2 := mcshapes.XYZ{X: origin.X + width - 1, Y: origin.Y + h, Z: origin.Z - 4}
			b := mcshapes.NewBox(mcshapes.WithCorner1(corner1), mcshapes.WithCorner2(corner2),
				mcshapes.WithSurface("minecraft:air"))
			b.Orient(direction)
			err = b.WriteShape(f)
			if err != nil {
				return fmt.Errorf("rm fall: %v", err)
			}
		}
	}

	return nil
}


// ClearForWall - clear space for a wal
// A wall in this context is meant to surround some area, such as a Mincraft village, and
// provides protection from Minecraft Hostile Mobs (zombies, creeper, spiders, ...). The wall
// should be at least 3 blocks high and it needs an overhang to keep the spiders out (spiders
// can crawl up a wall but cannot get past a ledge). The area inside the wall needs to be lit
// up so Hostile Mobs will not spawn. There needs to be clear space on the outside of the wall
// so the Hostile Mobs will not be able to jump to the top of the wall and thus into the secure
// area. Space is also left on the inside of the wall so villagers will not accidentally find
// their way outside the wall.
// The lava and water falls provide an excellent wall. Such falls are tall enough and come
// with a ledge on the outside. They are also visually stunning.
// This function clears space for the wall. The width, height, and depth parameters specify the
// extent of the cleared area. The wall is put in the middle of the cleared area. 
func ClearForWall(basepath string) error {
	origin := mcshapes.XYZ{X: 0, Y: 0, Z: -2}

	for _, direction := range []string{"north", "east", "south", "west"} {
		// Minecraft functions must have a suffix of ".mcfunction"
		fname := path.Join(basepath, "ClearForWall_"+direction) + ".mcfunction"
		f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("rm_fall open %v: %v", fname, err)
		}
		defer f.Close()

		// Minecraft will not accept a width that is too large. 150 is too large, 100 works.
		// Probably has something to do with the size of Minecraft chunks and how many chunks
		// are active and/or being visualized.
		width := 100
		height := 50
		depth := 17

		// Use a loop because Minecraft has a limit on total number of blocks per fill command.
		for h := -1; h <= height; h++ {
			corner1 := mcshapes.XYZ{X: origin.X, Y: origin.Y + h, Z: origin.Z}
			corner2 := mcshapes.XYZ{X: origin.X + width - 1, Y: origin.Y + h, Z: origin.Z - depth + 1}
			block_type := "air"
			if h == -1 {
				block_type = "sea_lantern"
			}
			b := mcshapes.NewBox(mcshapes.WithCorner1(corner1), mcshapes.WithCorner2(corner2),
				mcshapes.WithSurface("minecraft:"+block_type))
			b.Orient(direction)
			err = b.WriteShape(f)
			if err != nil {
				return fmt.Errorf("ClearForWall: %v", err)
			}
		}
	}

	return nil
}




// CreateSphere
func CreateSphere(basepath string, radius int, exteriorBlockType string,
	              interiorBlockType string) error {
	center := mcshapes.XYZ{X: radius, Y: 0, Z: radius+2}

	// Minecraft functions must have a suffix of ".mcfunction"
	srad := fmt.Sprintf("%d", radius)
	blkname := exteriorBlockType
	if interiorBlockType != "nothing" {
		blkname = exteriorBlockType + "_" + interiorBlockType
	}
	fname := basepath + "/Sphere_" + blkname + "_" + srad + ".mcfunction"
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("CreateSphere open %v: %v", fname, err)
	}
	defer f.Close()

	b := mcshapes.NewSphere(mcshapes.WithRadius(radius), mcshapes.WithCenter(center),
		mcshapes.WithSphereSurface("minecraft:" + exteriorBlockType),
		mcshapes.WithSphereInteriorSurface("minecraft:" + interiorBlockType))
	err = b.WriteShape(f)
	if err != nil {
		return fmt.Errorf("ClearForWall: %v", err)
	}

	return nil
}

