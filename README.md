# Prerequisite

- [Docker](https://docs.docker.com/get-docker/) installed
- [docker-compose](https://docs.docker.com/compose/install/) installed

# Start development server

First copy `.env.sample` to `.env` and modify environment variable.
For development environment the content in `.env.sample` should be working out of the box.

Then start our docker container with
```SH
$ docker-compose up
```

# Client Create Short URL

```
POST /shorten
```

| Parameter | Type | Description |
| --------- | ---- | ----------- |
| `url` | `string` | **Required**. URL to be shorten |
| `expiresIn` | `integer` | **Optional**. Expire duration in second |

## Response

API will return below response on success

```
{
  "url": string
}
```

API will return below response on error

```
{
  "error": [string]
}
```

# Authorization

All admin API endpoints required token based authorization. You can find API token from your `.env` file.

To authenticate an API request, you should provide your token in the Authorization header.

Alternatively, you may append the `token=[TOKEN]` as a GET parameter to authorize yourself to the API. But note that this is likely to leave traces in things like your history, if accessing the API through a browser.

```
GET /admin/shortUrls?token=12345
```

| Parameter | Type | Description |
| --------- | ---- | ----------- |
| `token` | `string` | **Required**. API token |

# Admin List URLs

```
GET /admin/shortUrls
```

| Parameter | Type | Description |
| --------- | ---- | ----------- |
| `offset` | `integer` | **Optional**. The position in which to start retrieve the records. Default 0 |
| `size` | `integer` | **Optional**. The number of result to return per request. Default 30 |
| `shortCode` | `string` | **Optional**. Short URL code to filter |
| `keyword` | `string` | **Optional**. Keyword to filter on domain name in full url |

## Response

API will return below response on success

```
{
  "data": [
    {
      "fullUrl": string,
      "code": string,
      "expiredAt": string, // Datetime format. Can be omit if empty
      "hitCount": integer
    }
  ],
  "totalCount": integer
}
```

API will return below response on error

```
{
  "error": [string]
}
```

# Admin Delete URL

```
DELETE /admin/shortUrls/{code}
```

| Parameter | Type | Description |
| --------- | ---- | ----------- |
| `code` | `string` | **Required**. Short URL code |

API will return `204` status on success and below response on error

```
{
  "error": [string]
}
```

# Status Codes

Shortening API will return below status codes:

| Status Code | Description |
| ----------- | ----------- |
| 200 | OK |
| 201 | Created |
| 204 | No content |
| 400 | Bad request |
| 403 | Forbidden |
| 404 | URL not found |
| 410 | Gone. URL was removed |
| 500 | Server error |
