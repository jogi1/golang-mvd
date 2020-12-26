# golang-mvd  

# building
golang 1.13 is required.

# output  
output is handled via javascript run in a vm. if a file "runme.js" is in the same dir as the parser it will be used instead of the inbuild default (wich can be found in "data/default.js")  
check ```examples/frame.js``` for an example of the ```on_frame(current, last, events)``` handling


