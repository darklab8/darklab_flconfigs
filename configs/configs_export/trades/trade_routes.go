package trades

import (
	"math"
	"strings"

	"github.com/darklab8/fl-configs/configs/configs_mapped"
	"github.com/darklab8/fl-configs/configs/configs_mapped/freelancer_mapped/data_mapped/universe_mapped/systems_mapped"
	"github.com/darklab8/fl-configs/configs/conftypes"
)

type SystemObject struct {
	nickname string
	pos      conftypes.Vector
}

func DistanceForVecs(Pos1 conftypes.Vector, Pos2 conftypes.Vector) float64 {
	// if _, ok := Pos1.X.GetValue(); !ok {
	// 	return 0, errors.New("no x")
	// }
	// if _, ok := Pos2.X.GetValue(); !ok {
	// 	return 0, errors.New("no x")
	// }

	x_dist := math.Pow((Pos1.X - Pos2.X), 2)
	y_dist := math.Pow((Pos1.Y - Pos2.Y), 2)
	z_dist := math.Pow((Pos1.Z - Pos2.Z), 2)
	distance := math.Pow((x_dist + y_dist + z_dist), 0.5)
	return distance
}

type WithFreighterPaths bool

/*
Algorithm should be like this:
We iterate through list of Systems:
Adding all bases, jump gates, jump holes, tradelanes as Vertexes.
We scan in advance nicknames for object on another side of jump gate/hole and add it as vertix
We calculcate distances between them. Distance between jump connections is 0 (or time to wait measured in distance)
We calculate distances between trade lanes as shorter than real distance for obvious reasons.
The matrix built on a fight run will be having connections between vertixes as hashmaps of possible edges? For optimized memory consumption in a sparse matrix.

Then on second run, knowing amount of vertixes
We build Floyd matrix? With allocating memory in bulk it should be rather rapid may be.
And run Floud algorithm.
Thus we have stuff calculated for distances between all possible trading locations. (edited)
[6:02 PM]
====
Then we build table of Bases as starting points.
And on click we show proffits of delivery to some location. With time of delivery. And profit per time.
[6:02 PM]
====
Optionally print sum of two best routes that can be started within close range from each other.
*/
func MapConfigsToFGraph(configs *configs_mapped.MappedConfigs, with_freighter_paths WithFreighterPaths) *GameGraph {
	graph := NewGameGraph()
	for _, system := range configs.Systems.Systems {

		var system_objects []SystemObject = make([]SystemObject, 0, 50)

		for _, system_obj := range system.Bases {
			object := SystemObject{
				nickname: system_obj.Base.Get(),
				pos:      system_obj.Pos.Get(),
			}

			for _, existing_object := range system_objects {
				distance := DistanceForVecs(object.pos, existing_object.pos)
				graph.SetEdge(object.nickname, existing_object.nickname, distance)
			}

			if strings.Contains(object.nickname, "proxy_") {
				continue
			}

			graph.vertex_to_calculate_paths_for[VertexName(object.nickname)] = true

			system_objects = append(system_objects, object)
		}

		for _, jumphole := range system.Jumpholes {
			object := SystemObject{
				nickname: jumphole.Nickname.Get(),
				pos:      jumphole.Pos.Get(),
			}

			jh_archetype := jumphole.Archetype.Get()

			// Condition is taken from FLCompanion
			// https://github.com/Corran-Raisu/FLCompanion/blob/021159e3b3a1b40188c93064f1db136780424ea9/Datas.cpp#L585
			// Check Aingar Fork for Disco version if necessary.
			if strings.Contains(jh_archetype, "_fighter") || strings.Contains(jh_archetype, "_notransport") || jh_archetype == "dsy_comsat_planetdock" {
				if !with_freighter_paths {
					continue
				}
			}

			for _, existing_object := range system_objects {
				distance := DistanceForVecs(object.pos, existing_object.pos)
				graph.SetEdge(object.nickname, existing_object.nickname, distance)
			}

			jumphole_target_hole := jumphole.GotoHole.Get()
			graph.SetEdge(object.nickname, jumphole_target_hole, 0)
			system_objects = append(system_objects, object)
		}

		for _, tradelane := range system.Tradelanes {
			object := SystemObject{
				nickname: tradelane.Nickname.Get(),
				pos:      tradelane.Pos.Get(),
			}

			next_tradelane, next_exists := tradelane.NextRing.GetValue()
			prev_tradelane, prev_exists := tradelane.PrevRing.GetValue()

			speed := 350
			tradelane_speed := 2250

			if next_exists && prev_exists {
				continue
			}

			// next or previous tradelane
			chained_tradelane := ""
			if next_exists {
				chained_tradelane = next_tradelane
			} else {
				chained_tradelane = prev_tradelane
			}
			var last_tradelane *systems_mapped.TradeLaneRing
			// iterate to last in a chain
			for {
				another_tradelane, ok := system.TradelaneByNick[chained_tradelane]
				if !ok {
					break
				}
				last_tradelane = another_tradelane

				if next_exists {
					chained_tradelane, _ = another_tradelane.NextRing.GetValue()
				} else {
					chained_tradelane, _ = another_tradelane.PrevRing.GetValue()
				}
				if chained_tradelane == "" {
					break
				}
			}

			if last_tradelane == nil {
				continue
			}

			distance := DistanceForVecs(object.pos, last_tradelane.Pos.Get())
			graph.SetEdge(object.nickname, last_tradelane.Nickname.Get(), distance*float64(speed)/float64(tradelane_speed))

			for _, existing_object := range system_objects {
				distance := DistanceForVecs(object.pos, existing_object.pos)
				graph.SetEdge(object.nickname, existing_object.nickname, distance)
			}

			system_objects = append(system_objects, object)
		}
	}
	return graph
}