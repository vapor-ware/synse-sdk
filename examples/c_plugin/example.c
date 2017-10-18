#include <stdlib.h>
#include <time.h>
#include "example.h"


// Example of a simple read - this just gets and returns a random
// number. The interface here is also kept pretty simple for this
// example, but could become more complicated depending on the needs
// of the actual protocol.
int
Read(int device, char* model)
{
    srand(time(NULL));
    return rand();
}
