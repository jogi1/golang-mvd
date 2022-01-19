/*
 * available functions:
 * __print(string): prints string to either stdout or the set outputfile, does not add a trailing newline
 * __println(string): same as __print but adds a newline
 * __SetOutputFile(filanem): sets the file where __print is redirected to
 * __SetOutputFile(filanem): sets the file where __print is redirected to
 * __StringSanatize(string): returns a sanatized version of the string
 * __StringSanatizeEscapes(string): returns a sanatized version of the string with escaped "\"
 * __StringConvertInt(string): returns a array of integers of the strings individual byted converted to its byte code
 */

// will before the demo is being parsed
function on_init(filename) {
    // filename - is the filaname the demo has
}

// will be called every demo frame
function on_frame(server, current_state, last_state, events, stats, fragmessages, players) {
    /*
     * server: server info
     * current_state: current frame state of the demo
     * last_state: last frame state of the demo
     * events: parser events this frame
     * stats: parser stats this frame
     * fragmessages: parsed fragmessages this frame
     * players: this frames aggregated info of players
     */
}

// will before the demo is finished parsing
function on_finish(server, current_state, stats, fragmessages, players, mod_parser_state) {
    /*
     * server: server info
     * current_state: final state of the demo
     * stats: final parser stats
     * fragmessages: all parsed fragmessages
     * players: final aggregated info of players
     * mod_parser_state: mod parser info
     */
}

