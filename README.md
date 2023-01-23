# Google Cloud Logging Data Source

## Overview

The Google Cloud Logging Data Source is a backend data source plugin for Grafana,
which allows users to query and visualize their Google Cloud logs in Grafana.

## Setup

### Download

Download this plugin to the machine Grafana is running on, either using `git clone` or simply downloading it as a ZIP file. For the purpose of this guide, we'll assume the user "alice" has downloaded it into their local directory "/Users/alice/grafana/". If you are running the Grafana server using a user such as `grafana`, make sure the user has access to the directory.

### Generate a JWT file

1.  if you don't have gcp project, add new gcp project. [link](https://cloud.google.com/resource-manager/docs/creating-managing-projects#console)
2.  Open the [Credentials](https://console.developers.google.com/apis/credentials) page in the Google API Console.
3.  Click **Create Credentials** then click **Service account**.
4.  On the Create service account page, enter the Service account details.
5.  On the `Create service account` page, fill in the `Service account details` and then click `Create`
6.  On the `Service account permissions` page, don't add a role to the service account. Just click `Continue`
7.  In the next step, click `Create Key`. Choose key type `JSON` and click `Create`. A JSON key file will be created and downloaded to your computer
8.  Note your `service account email` ex) *@*.iam.gserviceaccount.com
9.  Open the [Google Analytics API](https://console.cloud.google.com/apis/library/analytics.googleapis.com)  in API Library and enable access for your account
10. Open the [Google Analytics Reporting API](https://console.cloud.google.com/marketplace/product/google/analyticsreporting.googleapis.com?q=search&referrer=search&project=composed-apogee-307906)  in API Library and enable access for your GA Data

### Grafana Configuration

1. With Grafana restarted, navigate to `Configuration -> Data sources` (or the route `/datasources`)
2. Click "Add data source"
3. Select "Google Cloud Logging"
4. Provide credentials in a JWT file, either by using the file selector or pasting the contents of the file.
5. Click "Save & test" to test that logs can be queried from Cloud Logging.

## Licenses

Cloud Logging Logo (`src/img/logo.svg`) is from Google Cloud's [Official icons and sample diagrams](https://cloud.google.com/icons)

As commented, `JWTForm` and `JWTConfigEditor` are largely based on Apache-2.0 licensed [grafana-google-sdk-react](https://github.com/grafana/grafana-google-sdk-react/)
