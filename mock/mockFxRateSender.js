// USAGE:
// node mockFxRateSender.js <FX_RATE_ENDPOINT>

// Define the candidate's FX Rate Update endpoint
const FX_RATE_ENDPOINT = process.argv[2];

// Currencies: USD, EUR, JPY, GBP, AUD
const baseRates = {
  "USD/EUR": 1.085,
  "EUR/USD": 0.9217,
  "USD/JPY": 110.25,
  "JPY/USD": 0.0091,
  "USD/GBP": 0.75,
  "GBP/USD": 1.3333,
  "USD/AUD": 1.35,
  "AUD/USD": 0.7407,
  "EUR/JPY": 129.53,
  "JPY/EUR": 0.0077,
  "EUR/GBP": 0.85,
  "GBP/EUR": 1.1765,
  "EUR/AUD": 1.6,
  "AUD/EUR": 0.625,
  "GBP/JPY": 150.45,
  "JPY/GBP": 0.0066,
  "GBP/AUD": 1.8,
  "AUD/GBP": 0.5556,
  "AUD/JPY": 82.5,
  "JPY/AUD": 0.0121,
};

// Function to simulate rate changes, with a 1% chance of a 5% fluctuation
function getRandomRate(baseRate) {
  const isExtremeFluctuation = Math.random() < 0.05; // 5% chance

  if (isExtremeFluctuation) {
    // Apply a 5-10% fluctuation in either direction
    const fluctuation = (0.05 + Math.random() * 0.05) * (Math.random() < 0.5 ? -1 : 1);
    return (baseRate * (1 + fluctuation)).toFixed(4);
  } else {
    // Apply a standard fluctuation of +/- 0.5%
    const fluctuation = (Math.random() - 0.5) * 0.01;
    return (baseRate * (1 + fluctuation)).toFixed(4);
  }
}

// Function to send a mock FX rate update
async function sendFxRateUpdate() {
  // Randomly select a currency pair to update
  const pairs = Object.entries(baseRates);
  const [pair, rate] = pairs[Math.floor(Math.random() * pairs.length)];
  const updatedRate = getRandomRate(rate);

  // Build the mock FX rate payload
  const payload = {
    pair: pair,
    rate: updatedRate,
    timestamp: new Date().toISOString(),
  };

  // Send the rate update to the FX Rate Update endpoint
  await fetch(FX_RATE_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });

  console.log(
    `Sent FX rate update: ${pair} - ${updatedRate} at ${payload.timestamp}`
  );
}

// Schedule the function to run every 3-5 seconds
function main() {
  setInterval(() => {
    sendFxRateUpdate().catch((err) => {
      console.error("Error sending FX rate update:", err.message || err);
    });
  }, Math.floor(Math.random() * 2000) + 3000); // Random interval between 3 - 5 seconds, the frequency of the updates allows for a good amount of data to be sent
}

// Start the mock FX rate updates
main();
