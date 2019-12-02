# Vesnicobus

 A webservice providing an API that allows you to get information about selected buses in Prague Integrated Transport network (with respect to data that is actually available) and to estimate arrival to stops on their route. To achieve this, this application uses the following public APIs:

  - Golemio (Prague Open Data platform): https://golemio.cz/
  - Bing Maps Routes API: https://docs.microsoft.com/en-us/bingmaps/rest-services/routes/

## Setup
 1) Obtain a Golemio API key: https://api.golemio.cz/api-keys/auth/sign-in
 2) Obtain a Bing Maps API key: https://www.bingmapsportal.com/
 3) Setup a Redis server (preferably set it up to act as a LRU cache - 50MB of memory is enough, automatically remove least recently used keys, persistence can be disabled)
 6) Configure the vesnicobus service by updating values in ```vesnicobus.ini``` file accordingly, particularly the API keys and Redis connection
 7) Compile the go code
 8) Run the compiled binary (make sure the binary and the ini file are in the same directory)
 9) The server is ready

## API Documentation
https://app.swaggerhub.com/apis-docs/Silaedru/vesnicobus-service

## Related projects
[Vesnicobus Client](https://github.com/Silaedru/vesnicobus-client)