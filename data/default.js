function _p (indent_level, key, value) {
  m = ""
  for (x=0; x< indent_level; x++) {
    m = m + "\t"
  }
  return m + "\"" + key + "\": \"" + value + "\""
}

function on_finish () {
  print("{\n")
  print(_p(1, "hostname", sanatize(demo.Hostname)), ",\n")
  print(_p(1, "map_name", sanatize(demo.Mapname)), ",\n")
  print(_p(1, "map_file", sanatize(demo.Mapfile)),",\n")
  print("\t\"players\": [\n")
  var first = true 
  for (x in demo.Players) {
    p = demo.Players[x]
    if (p.Name.length == 0 || p.Spectator == true) {
      continue
    }
    if (first == false) {
      print(",\n")
    } else {
      first = false
    }
    print("\t\t{\n")
    print(_p(3, "name_sanatized", sanatize(p.Name)), ",\n")
    print(_p(3, "name_int", convert_int(p.Name)), ",\n")
    print(_p(3, "team_sanatized", sanatize(p.Team)), ",\n")
    print(_p(3, "team_int", convert_int(p.Team)), ",\n")
    print(_p(3, "frags", p.Frags), ",\n")
    print(_p(3, "deaths", p.Deaths), ",\n")
    print(_p(3, "Suicides", p.Suicides), ",\n")
    print("\t\t\t\"itemstats\": ")
    print(JSON.stringify(p.Itemstats) + "\n")
    print("\t\t}")
  }
  print("\n\t\]\n")
  print("}\n")


}

