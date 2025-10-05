// Include Inkplate library in the sketch
#include "Inkplate.h"

// ---------------- CHANGE HERE ---------------------:

extern const char *ssid;                       // Your WiFi SSID
extern const char *pass;                       // Your WiFi password
#define IMAGE_URL "http://example/screen.png"  // The URL of the image to display

#include "secret.h"  // Initialize the image URL and the WiFi credentials

// ---------------------------------------------------

// Create an object on Inkplate library and also set library into 3 Bit (grayscale) mode
Inkplate display(INKPLATE_3BIT);

#define HOUR_OF_REFRESH 3       // The hour of the day we refresh the screen at
#define TIMEZONE_OFFSET -5      // The timezone the screen is in

// Conversion factor for micro seconds to seconds
const uint64_t kMicrosToSecs = 1000000ULL;

void setup() {
  // Init serial communication
  Serial.begin(115200);
  Serial.println("Program started");

  // Init Inkplate library (you should call this function ONLY ONCE)
  display.begin();
  display.clearDisplay();

  display.setCursor(0, 0);
  display.setTextColor(0, 7);
  display.setTextSize(2);

  // Join wifi
  if (!display.connectWiFi(ssid, pass)) {
    display.printf("Wi-Fi failed ");
    Serial.println("Failed to connect to Wi-Fi");
    go_to_sleep(3600); // Sleep 1 hour before trying again.
    return;
  }
  Serial.println("Connected to Wi-Fi");

  // Draw an image on the screen
  display.drawImage(IMAGE_URL, display.PNG, 0, 0);

  // Print the battery charge on the screen.
  double voltage = display.readBattery();
  int battery_percentage = calc_battery_percentage(voltage);
  display.printf("battery %d%% ", battery_percentage);
  Serial.printf("Battery voltage %lf, percent charged %d%%\n", voltage, battery_percentage);

  // Get the current time; we don't care about daylight saving.
  struct tm timeinfo;
  if (!display.getNTPDateTime(&timeinfo, TIMEZONE_OFFSET)) {
    display.printf("NTP failed ");
    Serial.println("Failed to get time from NTP");
    go_to_sleep(2 * 3600); // Sleep 2 hours beforing trying again.
    return;
  }
  display.printf("at %02d:%02d:%02d ", timeinfo.tm_hour, timeinfo.tm_min, timeinfo.tm_sec);
  Serial.printf("Current time is %02d:%02d:%02d\n",
              timeinfo.tm_hour, timeinfo.tm_min, timeinfo.tm_sec);

  // Figure out the number of seconds since midnight for the current time, and the refresh.
  int current_secs = timeinfo.tm_hour * 3600 + timeinfo.tm_min * 60 + timeinfo.tm_sec;
  int refresh_secs = HOUR_OF_REFRESH * 3600;

  uint64_t secs_until_refresh;
  if (current_secs < refresh_secs) {
    // Refresh is later today
    secs_until_refresh = refresh_secs - current_secs;
  } else {
    // Refresh is tomorrow
    secs_until_refresh = (24 * 3600 - current_secs) + refresh_secs;
  }

  go_to_sleep(secs_until_refresh);
}

int calc_battery_percentage(double voltage) {
  const double VOLT_MIN = 3.2;
  const double VOLT_MAX = 4.1;
  double normalized = (voltage - VOLT_MIN) / (VOLT_MAX - VOLT_MIN);
  int percentage = (int)(normalized * 100 + 0.5);

  if (percentage < 0) return 0;
  if (percentage > 100) return 100;
  return percentage;
}

void go_to_sleep(uint64_t duration_secs) {
  display.printf("sleeping for %.2fh ", (float)duration_secs/3600.0);
  display.display();

  Serial.printf("Going to sleep for %.2f hours\n", (float)duration_secs/3600.0);
  esp_sleep_enable_timer_wakeup(duration_secs * kMicrosToSecs);
  esp_sleep_enable_ext0_wakeup(GPIO_NUM_36, 0);
  esp_deep_sleep_start();
}

void loop() {
  // No-op. At the end of setup(), we go into deep sleep, so we never reach this.
}
