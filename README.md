# Conduit Connector Zendesk

## Source

The Zendesk connector will connect to the Zendesk API through the `url` constructed and start fetching the record until `end_of_stream` is true. 

The CDC mode for the source connector is constructed based on the incremental export flow of Zendesk. 

### Incremental Flow Zendesk

Incremental Export API of Zendesk used to get items that are created or changed since the last request.
The maximum request for the Incremental API endpoint is restricted to 10 request per minute.

The `rate_limit_error` handled by Zendesk by throwing `http_response_code` 429 and `Retry_After` is set to 93seconds.

### Pagination
Cursor based incremental exports

Cursor based Pagination format

|   name          |   Type    |  Comment                |
|-------------------------------------------------------------------------------------------------------------------
| `after_url`     |   string  |  URL to fetch the next page of results
| `after_cursor`  |   string  |  Cursor to fetch the next page of results
| `before_url`    |   string  |  URL to fetch the previous page of results. If no previous page, value is null
| `before_cursor` |   string  |  Cursor to fetch the previous page of results. If no previous page, value is null      


Common JSON attribute added to the response

``` json
{
    "end_of_stream": true
}
```
### Initial Cursor Request 

- https://testlab.zendesk.com/api/v2/incremental/tickets/cursor.json?start_time=1532034771 

`start_time` is arbitary time used only once in the cursor based exports. Subsequent pages and exports handled by `after_cursor`

`after_url` - Position in sdk.Record to fetch if the pipeline is paused or `end_of_stream` is false
- https://testlab.zendesk.com/api/v2/incremental/tickets/cursor.json?cursor=MTY1MDI3NzcyMS4wfHwxNXw%3D"

### Configuration

| name                  | description                                                                  | required | default |
| -------               | ---------------------------------------------------------------------------  | -------- | ------- |
|`zendesk.domain`       | domain is the registered by organization to zendesk                          | true     |         |
|`zendesk.username`     | username is the registered for login                                         | true     |         |
|`zendesk.apitoken`     | password associated with the username for login                              | true     |         |
|`pollingPeriod`        | pollingPeriod is the frequency of conduit hitting zendesk API- Default is 2m | false    |  "2m"   |

##### NOTE: `pollingPeriod` will be in time.Duration - `2ns`,`2ms`,`2s`,`2m`,`2h`

## Destination
The destination Zendesk connector will connect to the api using `zendesk.domain`, `zendesk.username`,`zendesk.apitoken`. Communicate to the `Configure` and if succcess, it will pass control to `Open` else throws the error back. Once the zendesk client is initialized. On failing connector is not ready to write it to zendesk.

### WriteAsync
The source input from server will be written in the `buffer`, size of the buffer is specified in the configuration. Once the buffer if full it write the record to zendesk bulk import api `create_many`. Each object from tickets array is unmarshalled and made comptabile to write it to destination zendesk.

### Configuration
| name                  | description                                                                  | required | default |
| -------               | ---------------------------------------------------------------------------  | -------- | ------- |
|`zendesk.domain`       | domain is the registered by organization to zendesk                          | true     |         |
|`zendesk.username`     | username is the registered for login                                         | true     |         |
|`zendesk.apitoken`     | password associated with the username for login                              | true     |         | 

# Limitations

- IncrementalExport will take maximum 10 API request per minute
- `per_page` is a optional parameter for default per page result. default is set to 1000
- `bufferSize` is set to 100, as bulk import moves 100 tickets max in one request
- Ticket import can be authorized only by `admins`

# References

- https://developer.zendesk.com/documentation/ticketing/using-the-zendesk-apibest-practices-for-avoiding-rate-limiting/#catching-errors-caused-by-rate-limiting
- https://developer.zendesk.com/api-reference/ticketing/ticket-management/incremental_exports/#cursor-based-pagination-json-format
- https://developer.zendesk.com/documentation/ticketing/managing-tickets/using-the-incremental-export-api/#cursor-based-incremental-exports
- https://developer.zendesk.com/api-reference/ticketing/tickets/ticket_import/#ticket-bulk-import