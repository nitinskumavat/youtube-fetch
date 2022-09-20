# Youtube fetcher
This  app preiodically fetch the data from the youtube after certain interval and push the same data in mongo db.  It supports for multiple keys based on round robin approach. If Quota exceeded for first key it will use the next key. Api keys value can be found in Dockerfile

## Tech stack
- Golang
- Mongo DB 

## Apis
1.  `GET`  ` /videos?page=1` Returns the list of videos data in reverse chronological order. This endpoint returns the 10 items per page. **page=1** is the page number. For sorting published_at is added as index.

2. `GET`  `/search?query=mobile` Returns the list of videos with the matched query in the title or description of video data. The query will  output the match with max_edit 1.  **Example** `/search?query=niti ` will also match with **niki** or **kiti** in the title or description.
Search index created on title and description in MongoDb Atlas Search for text search.


## Running the docker
1. Build docker image `docker build -t youtube-api . `
2. Run docker image `docker run -d -p 12345:12345 youtube-api` this will run docker app server on port 12345 .

## Imporvements to be done
- Add pagination depending on Id . Instead of skipping the records using page number (logic implemented but commented bit more research required).
- Add logger.

