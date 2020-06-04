function _p (indent_level, key, value) {
  m = ""
  for (x=0; x< indent_level; x++) {
    m = m + "\t"
  }
  return m + "\"" + key + "\": \"" + value + "\""
}

function on_finish (state, stats, demo, server) {
  print("{\n")
  print(_p(1, "hostname", sanatize(server.Hostname)), ",\n")
  print(_p(1, "map_name", sanatize(server.Mapname)), ",\n")
  print(_p(1, "map_file", sanatize(demo.Modellist[0])),",\n")
  print("\t\"players\": [\n")
  var first = true 
  for (index in state.Players) {
    p = state.Players[index]
    if (p.Name.length == 0 || p.Spectator == true) {
      continue
    }
    if (first == false) {
      print(",\n")
    } else {
      first = false
    }
    print("\t\t{\n")
    print(_p(3, "name", p.Name), ",\n")
    print(_p(3, "name_sanatized", sanatize(p.Name)), ",\n")
    print(_p(3, "name_int", convert_int(p.Name)), ",\n")
    print(_p(3, "team", p.Team), ",\n")
    print(_p(3, "team_sanatized", sanatize(p.Team)), ",\n")
    print(_p(3, "team_int", convert_int(p.Team)), ",\n")
    print(_p(3, "frags", p.Frags), ",\n")
    print("\t\t\t\"stats\": ")
    stat = stats[index]
    print(JSON.stringify(stat) + "\n")
    print("\t\t}")
  }
  print("\n\t\]\n")
  print("}\n")
}

