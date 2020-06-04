function on_frame(current, last, events, stats, server) {
  for (e in events) {
    e = events[e]
    if (e.Type == event_types.death) {
      print(current.Time + ": " + current.Players[e.Player_Number].Name + " died\n")
    } else if (e.Type == event_types.drop) {
      print(current.Time + ": " + current.Players[e.Player_Number].Name + " has dropped " + items_name[e.Item_Type] + "\n")
    } else if (e.Type == event_types.pickup) {
      print(current.Time + ": " + current.Players[e.Player_Number].Name + " has picked up " + items_name[e.Item_Type] + "\n")
    } else if (e.Type == event_types.spawn) {
      print(current.Time + ": " + current.Players[e.Player_Number].Name + " has spawned\n")
    } else {
      print(e.Type + "\n")
    }
  }
}
