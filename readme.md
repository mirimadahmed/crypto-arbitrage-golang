# Crypto Arbitrage Detector

This Go program detects arbitrage opportunities between Bybit and Binance cryptocurrency exchanges.

## Description

The Crypto Arbitrage Detector fetches real-time price data from Bybit and Binance exchanges, compares the prices for matching pairs, and identifies potential arbitrage opportunities. It considers transaction fees and allows you to set a minimum profit threshold.

## Features

- Fetches real-time price data from Bybit and Binance
- Compares prices for matching pairs across both exchanges
- Considers transaction fees in calculations
- Configurable minimum profit threshold
- Detailed logging of the comparison process
- Sample output for debugging when no opportunities are found

## Prerequisites

- Go 1.15 or higher
- github.com/shopspring/decimal package

## Installation

1. Clone this repository:
   ```
   git clone https://github.com/mirimadahmed/crypto-arbitrage-golang.git
   ```

2. Navigate to the project directory:
   ```
   cd crypto-arbitrage-golang
   ```

3. Install the required package:
   ```
   go get github.com/shopspring/decimal
   ```

## Usage

Run the program with:
go run main.go


## Configuration

You can adjust the following constants in the `main.go` file:

- `minProfitPercentage`: Minimum profit percentage to consider as an arbitrage opportunity (default: 0.02 or 2%)
- `transactionFee`: Transaction fee per exchange (default: 0.001 or 0.1%)

## Output

The program will output:

- Number of pairs retrieved from each exchange
- Number of pairs compared
- Detailed information about any arbitrage opportunities found
- If no opportunities are found, sample comparisons for debugging

## Disclaimer

This program is for educational purposes only. Cryptocurrency trading carries a high level of risk, and there is always the potential for loss. The results of this program are not guaranteed. Always do your own research before making any investment decisions.

## License

[MIT License](LICENSE)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## Support

If you encounter any problems or have any questions, please open an issue in this repository.
