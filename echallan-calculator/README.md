# eChallan Calculator

eChallan Calculator is a microservice responsible for calculation rules of the eChallan system.

## API Routes
- `POST /echallan-calculator/v1/_calculate`: Calculates the tax heads and creates a demand based on the calculation criteria.
- `POST /echallan-calculator/v1/_getbill`: Retrieves the calculated bill for a specific challan.

## Math Engine Responsibilities
The math engine handles calculating tax amounts according to business rules. It:
- Maps challan amounts to tax head estimates
- Sums the amounts to provide total estimates
- Interfaces with the billing service to generate demands and cancel bills
