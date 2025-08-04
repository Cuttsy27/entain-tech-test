# Entain Technical Test - Solution Overview

## Preamble - A bit about me

Hi! My name is Shane. Firstly I would like to say that this has been a fun project to work on over the weekend - give my praise to the creator! I mainly use Typescript and C#, so coming to Go has been a great learning experience. I enjoy the simplicity of the language while it is still fully-featured in that I can easily write tests, run the app, build the app etc all without additional config (huge difference to JS/Node in particular).
I was given this task late on Friday afternoon to try and speed the process along, hoping I can have it submitted before lining up the initial interview. I worked hard to get the project completed as fast as possible.

## Introduction

This document provides an overview of my contributions to the technical test, where I extended an existing microservices application for retrieving racing and sports data.

## Application Architecture

The application is structured using a microservices architecture, which I worked within and extended. The existing components were:

*   **API Gateway (`api`):** A RESTful gateway that serves as the public-facing entry point, which I updated to route requests to the new `sports` service.
*   **Racing Service (`racing`):** An existing gRPC service for racing data, which I enhanced with several new features.

As part of the test, I also built a new service based on the existing racing service:

*   **Sports Service (`sports`):** A new gRPC service I created to manage and provide data for various sports events.

This design separates concerns, and my work continues that pattern, allowing for independent development and deployment of each service.

## Implemented Features & Contributions

Here are the features and enhancements I implemented as part of the technical test:

### 1. Enhanced the List RPC Filtering feature

I added more advanced filtering and sorting capabilities to the existing `racing` service:

*   **Visibility Filter:** I introduced a filter that allows you to fetch only the races that are marked as `visible`. This is useful for showing only the races that are currently active or relevant to users.

#### Discussion
I added an `optional bool visible` field for the filter because both true and false are valid values for the visible property. Now we can first check if the visible filter is applied (not nil) instead of defaulting to false. The user can get races unfiltered, or get races where visible = true or visible = false.

### 2. Enhanced the List RPC further with an OrderBy feature

*   **Flexible Sorting:** I implemented customizable sorting. You can now specify which field to sort by and in what direction (ascending or descending). By default, the results are sorted by their `advertised_start_time`.

#### Discussion
I added the order_by string field to the ListRacesRequest message. It is a string as recommended by [Google AIP-132](https://google.aip.dev/132#ordering), though I am still applying the field to the Request message which is then provided by the user as part of the POST request body, which is not recommended by the AIP documentation (though it is consistent with the existing codebase).

### 3. Real-time Race and Event Status

I added a `status` field to the `Race` resource. This field is derived from `advertised_start_time` on each request:

*   **`OPEN`:** If the `advertised_start_time` is in the future, the status will be `OPEN`.
*   **`CLOSED`:** If the `advertised_start_time` is in the past, the status will be `CLOSED`.

This gives you a real-time view of the status of each race.

#### Discussion
There may be room for improvement around how close an advertised_start_time can be to the current time during comparison because by the time the result gets back to the client, the status might be OPEN when it should now be CLOSED (though this is an issue for the entire list of races if it is not constantly updated in real-time).

### 4. Direct Resource Retrieval

I implemented a `GetRace` RPC to allow you to fetch a single race by its ID, providing a direct and efficient way to access specific race resources.

#### Discussion
This one was pretty straight-forward. I followed the [AIP-131](https://google.aip.dev/131) docs and created the RPC as a GET request, structured hierarchically by resource like `/v1/races/{id}` (although the original List RPC uses `/v1/list-races` as a POST request). If this were a real application, I would suggest refactoring the List RPC to use the `/races` path instead, with a resulting `/v2/races` GET request signature.

### 5. New `sports` Service

I built the new `sports` service from scratch to handle sports event data. This service provides similar functionality to the `racing` service, with `ListEvents` and `GetEvent` RPCs. The `Event` resource includes an `id`, `name`, `sport`, `visible` and `advertised_start_time`, and it also has the same dynamic `status` field as the `racing` service.

#### Discussion
The new sports service was the largest and most complex task for me. Looking back at it, I may have worked too strictly within the bounds of the Google AIP recommendations and created more complexity for myself and for the codebase as a whole. Google [AIP-132](https://google.aip.dev/132#filtering) recommends providing a filter as a string (same as order_by) and because the List RPC should use a GET request without a body, the filter and order_by strings are provided to the application as query parameters. 

Working with strings for both the filter and order_by increased complexity in their respective repo methods (applyFilter, applyOrderBy) because the strings need to be validated with regexp, split and checked for allowed properties, then build the SQL query. This method seems to be more prone to SQL injection and bugs, and is more difficult to extend. Perhaps there are better ways to implement this that still align with Google AIP recommendations. Additionally, while maybe not following AIP guidelines so strictly, the existing filtering in the racing service had an improved DX due to simpler code and better type-safety (importing the `ListRacesRequestFilter` type from proto) and was less prone to bugs and SQL injection.

In hindsight, it may have been better for me to stay consistent with the existing codebase, which is probably what I would have done if given a ticket on a real-world project, but questioned or created tasks for refactoring to better align with AIP in the future.

## Code Improvements and Best Practices

Throughout the project, I focused on writing clean, maintainable code and following best practices:

*   **Google AIP Design Standards:** I designed the new APIs to be consistent with the Google AIP Design Guide, using standard methods like `List` and `Get` and a resource-oriented approach.
*   **Clear and Concise Code:** I made sure my contributions are well-documented to make them easy to read and understand.
*   **Robust Error Handling:** I implemented error handling in the new RPCs to gracefully handle any invalid requests or other issues that may arise.
*   **Dependency Management:** I worked within the existing Go modules setup to manage dependencies, ensuring that the builds remain reproducible.

## Getting Started

Here's how you can get the application up and running:

1.  **Start the Services:**
    *   Open a terminal, navigate to the `racing` directory, and run:
        ```bash
        go build && ./racing
        ```
    *   In another terminal, navigate to the `sports` directory, and run:
        ```bash
        go build && ./sports
        ```
    *   In a third terminal, navigate to the `api` directory, and run:
        ```bash
        go build && ./api
        ```

2.  **Make API Requests:**
    *   **List Races:**
        ```bash
        curl -X "POST" "http://localhost:8000/v1/list-races" -H 'Content-Type: application/json' -d '{}'
        ```
    *   **List Races (Visible Only):**
        ```bash
        curl -X "POST" "http://localhost:8000/v1/list-races" -H 'Content-Type: application/json' -d '{"filter": {"visible": true}}'
        ```
    *   **Get Race by ID:**
        ```bash
        curl "http://localhost:8000/v1/races/{id}"
        ```
    *   **List Sports Events:**
        ```bash
        curl "http://localhost:8000/v1/events"
        ```
    *   **List Sports Events (Soccer and Basketball AND Visible only), order by advertised_start_time in descending order:**
        First example for readability only
        ```bash
        curl "http://localhost:8000/v1/events?filter=sport in (\"Soccer\", \"Basketball\") AND visible = true&order_by=advertised_start_time desc"
        ```
        ```bash
        curl "http://localhost:8000/v1/events?filter=sport%20in%20(%22Soccer%22,%20%22Basketball%22)%20AND%20visible%20=%20true&order_by=advertised_start_time%20desc"
        ```
    *   **Get Sports Event by ID:**
        ```bash
        curl "http://localhost:8000/v1/events/{id}"
        ```