> ## Documentation Index
> Fetch the complete documentation index at: https://docs.withterminal.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Lucid ELD

> This document provides an overview of Terminal's integration with Lucid ELD. It shows which data models are currently available in production and which ones are in active development.

## Provider Details

* **Provider Code**: `lucid-eld`
* **Available History**: 180 days
* **Vehicle Location Sample Rate**: 60 seconds
* **[Crash Reports](/terminal-platform/crash-reports)**: *Not supported*

## Supported Models

### Entities

<AccordionGroup>
  <Accordion title="Device" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Driver" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#45a049" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM369 209L241 337c-9.4 9.4-24.6 9.4-33.9 0l-64-64c-9.4-9.4-9.4-24.6 0-33.9s24.6-9.4 33.9 0l47 47L335 175c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9z"/></svg>}>
    The percentage of connections that have data populated for each field, averaged across the connections using this provider.

    | Field            | Coverage                                              |
    | ---------------- | ----------------------------------------------------- |
    | `email`          | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `firstName`      | <Badge color="green" icon="circle-check">100%</Badge> |
    | `groups`         | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `id`             | <Badge color="green" icon="circle-check">100%</Badge> |
    | `lastName`       | <Badge color="green" icon="circle-check">100%</Badge> |
    | `license.number` | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `license.state`  | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `middleName`     | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `phone`          | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `provider`       | <Badge color="green" icon="circle-check">100%</Badge> |
    | `sourceId`       | <Badge color="green" icon="circle-check">100%</Badge> |
    | `status`         | <Badge color="green" icon="circle-check">100%</Badge> |
    | `username`       | <Badge color="green" icon="circle-check">100%</Badge> |
  </Accordion>

  <Accordion title="Group" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Trailer" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Vehicle" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#45a049" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM369 209L241 337c-9.4 9.4-24.6 9.4-33.9 0l-64-64c-9.4-9.4-9.4-24.6 0-33.9s24.6-9.4 33.9 0l47 47L335 175c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9z"/></svg>}>
    The percentage of connections that have data populated for each field, averaged across the connections using this provider.

    | Field                 | Coverage                                              |
    | --------------------- | ----------------------------------------------------- |
    | `devices`             | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `fuelTankCapacity`    | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `fuelType`            | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `groups`              | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `id`                  | <Badge color="green" icon="circle-check">100%</Badge> |
    | `licensePlate.number` | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `licensePlate.state`  | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `make`                | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `model`               | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `name`                | <Badge color="green" icon="circle-check">100%</Badge> |
    | `provider`            | <Badge color="green" icon="circle-check">100%</Badge> |
    | `sourceId`            | <Badge color="green" icon="circle-check">100%</Badge> |
    | `status`              | <Badge color="green" icon="circle-check">100%</Badge> |
    | `vin`                 | <Badge color="green" icon="circle-check">100%</Badge> |
    | `year`                | <Badge color="red" icon="circle-xmark">0%</Badge>     |
  </Accordion>
</AccordionGroup>

### Real-Time Data

<AccordionGroup>
  <Accordion title="HOS Available Time" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Latest Trailer Location" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Latest Vehicle Location" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#45a049" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM369 209L241 337c-9.4 9.4-24.6 9.4-33.9 0l-64-64c-9.4-9.4-9.4-24.6 0-33.9s24.6-9.4 33.9 0l47 47L335 175c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9z"/></svg>}>
    Supported by Lucid ELD. Attribute coverage is not yet tracked for this model.
  </Accordion>
</AccordionGroup>

### Historical Data

<AccordionGroup>
  <Accordion title="Fault Code Event" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="HOS Daily Log" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="HOS Log" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="IFTA Vehicle Month" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Safety Event" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Trip" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Vehicle Location" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#45a049" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM369 209L241 337c-9.4 9.4-24.6 9.4-33.9 0l-64-64c-9.4-9.4-9.4-24.6 0-33.9s24.6-9.4 33.9 0l47 47L335 175c9.4-9.4 24.6-9.4 33.9 0s9.4 24.6 0 33.9z"/></svg>}>
    The percentage of connections that have data populated for each field, averaged across the connections using this provider.

    | Field               | Coverage                                              |
    | ------------------- | ----------------------------------------------------- |
    | `address.formatted` | <Badge color="green" icon="circle-check">100%</Badge> |
    | `driver`            | <Badge color="red" icon="circle-xmark">0%</Badge>     |
    | `heading`           | <Badge color="green" icon="circle-check">100%</Badge> |
    | `id`                | <Badge color="green" icon="circle-check">100%</Badge> |
    | `locatedAt`         | <Badge color="green" icon="circle-check">100%</Badge> |
    | `location`          | <Badge color="green" icon="circle-check">100%</Badge> |
    | `provider`          | <Badge color="green" icon="circle-check">100%</Badge> |
    | `speed`             | <Badge color="green" icon="circle-check">100%</Badge> |
    | `vehicle`           | <Badge color="green" icon="circle-check">100%</Badge> |
  </Accordion>

  <Accordion title="Vehicle Stat Log" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>

  <Accordion title="Vehicle Utilization" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>
</AccordionGroup>

### Other

<AccordionGroup>
  <Accordion title="Camera Media" icon={<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" height="16" width="16"><path fill="#e6a800" d="M256 512A256 256 0 1 0 256 0a256 256 0 1 0 0 512zM216 336h24V272H216c-13.3 0-24-10.7-24-24s10.7-24 24-24h48c13.3 0 24 10.7 24 24v88h8c13.3 0 24 10.7 24 24s-10.7 24-24 24H216c-13.3 0-24-10.7-24-24s10.7-24 24-24zm40-208a32 32 0 1 1 0 64 32 32 0 1 1 0-64z"/></svg>}>
    Terminal does not support this model for Lucid ELD yet.
  </Accordion>
</AccordionGroup>

## References

<CardGroup cols={2}>
  <Card title="Lucid ELD Developer Documentation" href="https://api.drivehos.app/partner/swagger" arrow="true" horizontal />

  <Card title="Lucid ELD Login" href="https://app.lucideld.com/" arrow="true" horizontal />
</CardGroup>
