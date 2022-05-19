# Conduit Connector Zendesk

## Source Connector

The Zendesk client connector will connect with Zendesk API through the `url` constructed using subdomain specific to individual organization. Upon successful configuration with `zendesk.userName ` and `zendesk.apiToken` the tickets from the given domain is fetched until `end_of_stream` is true.  

The CDC mode for the source connector is constructed based on the incremental export flow of Zendesk and scope is restricted to zendesk tickets. 

### Generating API token in Zendesk
The apiToken for the zendesk can be created throguh zendesk portal by login as admin.
- https://support.zendesk.com/hc/en-us/articles/4408889192858-Generating-a-new-API-token#topic_bsw_lfg_mmb

### Configuration - Source

| name                  | description                                                                  | required | default |
| -------               | ---------------------------------------------------------------------------  | -------- | ------- |
|`zendesk.domain`       | domain is the registered by organization to zendesk                          | true     |         |
|`zendesk.userName`     | username is the registered for login                                         | true     |         |
|`zendesk.apiToken`     | password associated with the username for login                              | true     |         |
|`pollingPeriod`        | pollingPeriod is the frequency of conduit hitting zendesk API- Default is 2m | false    |  "2m"   |

##### NOTE: `pollingPeriod` will be in time.Duration - `2ns`,`2ms`,`2s`,`2m`,`2h`

### Initial Cursor Request 

- https://testlab.zendesk.com/api/v2/incremental/tickets/cursor.json?start_time=1532034771 

`start_time` is arbitary time used only once in the cursor based exports. Subsequent pages and exports handled by `after_cursor`

`after_url` - Position in sdk.Record to fetch if the pipeline is paused or `end_of_stream` is false
- https://testlab.zendesk.com/api/v2/incremental/tickets/cursor.json?cursor=MTY1MDI3NzcyMS4wfHwxNXw%3D"

### Incremental Flow Zendesk

Incremental Export API of Zendesk used to get items that are created or changed since the last request.
The maximum request for the Incremental API endpoint is restricted to 10 request per minute. Exceeding the limit, zendesk client receives `rate_limit_error` from zendesk response header (httpsStatusCode- 429), along with status code it will send the cool-down period as `Retry-After` parameter by Zendesk. The cool-down period is configurable time given by zendesk and subjective to latency with zendesk api server.

### Pagination

#### Cursor based Pagination format 

|   name            |   Type        |  Comment                                                                        |
|----------------     -------------   --------------------------------------------------------------------------------
| `after_url`       |   string      |  URL to fetch the next page of results
| `after_cursor`    |   string      |  Cursor to fetch the next page of results
| `before_url`      |   string      |  URL to fetch the previous page of results. If no previous page, value is null
| `before_cursor`   |   string      |  Cursor to fetch the previous page of results. If no previous page, value is null   

NOTE: Pagination limit is determined by `after_cursor` given by zendesk, refer https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/#cursor-based-pagination-json-format 

Common JSON attribute added to the response

``` json
{
    "end_of_stream": true
}
```

# Limitations
- IncrementalExport will take maximum 10 API request per minute. On reaching the limit, zendesk api will throw `Retry-After`, cool-down time provided by zendesk that will prevent client to request to zendesk api server.  
- If response code 429 is given then total duration will sum of `Retry-After` and `pollingPeriod`  time.After()
- `per_page` is a optional parameter for default per page result. default is set to 1000

## Destination Connector
The destination connector import the ticket from the conduit as individua ticket object and store it to the buffer as array of tickets. Once the maxBufferSize is reached it will push the tickets to zendesk using bulk import api. Configuration for zendesk destination api includes `zendesk.domain`, `zendesk.userName`,`zendesk.apiToken`. Communication established using `Configure` method and if succcess, it will pass control to `Open` else throws the error back. Once the zendesk client is initialized, it will wait for the buffer to write the data to zendesk destination account. On failing connector is not ready to write it to zendesk.

### Configuration - Destination
| name                  | description                                                                  | required | default |
| -------               | ---------------------------------------------------------------------------  | -------- | ------- |
|`zendesk.domain`       | domain is the registered by organization to zendesk                          | true     |         |
|`zendesk.userName`     | username is the registered for login                                         | true     |         |
|`zendesk.apiToken`     | password associated with the username for login                              | true     |         |
|`bufferSize`           | bufferSize stores the ticket objects as array                                | false    |   100   |

### WriteAsync
The source input from server will be written in the `buffer`, size of the buffer is specified in the configuration. Once the buffer if full it write the record to zendesk bulk import api `create_many`. Each object from tickets array is unmarshalled and made comptabile to write it to destination zendesk.

# Limitations
- default `bufferSize` is set to 100, as bulk import moves 100 tickets max in one request
- Ticket import can be authorized only by `admins`
- Import for zendesk is scoped only with tickets

# References

- https://developer.zendesk.com/documentation/ticketing/using-the-zendesk-apibest-practices-for-avoiding-rate-limiting/#catching-errors-caused-by-rate-limiting
- https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/#cursor-based-pagination-json-format
- https://developer.zendesk.com/documentation/ticketing/managing-tickets/using-the-incremental-export-api/#cursor-based-incremental-exports
- https://developer.zendesk.com/api-reference/ticketing/tickets/ticket_import/#ticket-bulk-import