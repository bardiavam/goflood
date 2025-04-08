# GoFlood - Proxy-enabled HTTP/HTTPS Flood Testing Tool

GoFlood is a security research tool for testing web infrastructure resilience against distributed denial-of-service (DDoS) attacks. It provides a framework to simulate HTTP/HTTPS flood attacks using proxy servers.

## ⚠️ IMPORTANT DISCLAIMER

This tool is provided **FOR SECURITY RESEARCH PURPOSES ONLY**. Using this tool against any website or service without explicit permission from the owner is:

1. Likely **ILLEGAL** in most jurisdictions
2. A violation of most service providers' terms of service
3. **UNETHICAL** and potentially harmful to legitimate users

**THE AUTHORS TAKE NO RESPONSIBILITY** for any misuse of this software. Users are solely responsible for ensuring they have proper authorization before conducting any security testing.

## Features

- HTTP/HTTPS flooding capability through proxies
- Configurable concurrency level (number of workers)
- Proxy validation before attack
- Automatic proxy rotation
- Detailed statistics reporting
- Graceful shutdown support
- Customizable request parameters
- Automatic proxy grabber from public sources

## Installation

```bash
# Clone the repository (if you're using git)
git clone https://github.com/bardiavam/goflood.git
cd goflood

# Build the executable
go build -o goflood

# Alternative: Install directly with Go
go install github.com/bardiavam/goflood@latest
```

## Usage

```bash
# Basic usage
./goflood -target http://example.com -proxies proxies.txt -duration 30s

# Advanced usage with all options
./goflood -target https://example.com -proxies proxies.txt -workers 1000 -duration 2m -proxycheck http://google.com -skipcheck -verbose

# Using the proxy grabber
./goflood -target http://example.com -grabproxies -graboutput my_proxies.txt -duration 1m
```

### Required Flags

- `-target`: The target URL to attack (must include http:// or https://)
- `-proxies` or `-grabproxies`: Either specify a file with proxies or use the grabber to fetch them

### Optional Flags

- `-workers`: Number of concurrent workers (default: 500)
- `-duration`: Duration of the attack (default: 60s, format: 30s, 5m, 2h)
- `-proxycheck`: URL to use for proxy checking (default: Google's generate_204 endpoint)
- `-skipcheck`: Skip the proxy validation phase (use all proxies regardless of whether they work)
- `-verbose`: Enable verbose logging during the attack

### Proxy Grabber Options

- `-grabproxies`: Enable automatic proxy grabbing from public sources
- `-graboutput`: Filename to save grabbed proxies (default: proxies.txt)

## Contact

For questions, suggestions, or contributions, please contact:

- Name: Bardia
- GitHub: [@bardiavam](https://github.com/bardiavam)
- Email: bardiawam@gmail.com

# GoFlood - ابزار تست فلود HTTP/HTTPS با پروکسی

GoFlood یک ابزار تحقیقاتی امنیتی برای آزمایش مقاومت زیرساخت‌های وب در برابر حملات توزیع‌شده منع سرویس (DDoS) است. این ابزار چارچوبی برای شبیه‌سازی حملات فلود HTTP/HTTPS با استفاده از سرورهای پروکسی فراهم می‌کند.

## ⚠️ سلب مسئولیت مهم

این ابزار **فقط برای اهداف تحقیقاتی امنیتی** ارائه شده است. استفاده از این ابزار علیه هر وب‌سایت یا سرویس بدون اجازه صریح از مالک آن:

1. احتمالاً در اکثر حوزه‌های قضایی **غیرقانونی** است
2. نقض شرایط خدمات اکثر ارائه‌دهندگان سرویس است
3. **غیراخلاقی** و بالقوه مضر برای کاربران مشروع است

**نویسندگان هیچ مسئولیتی** در قبال سوء استفاده از این نرم‌افزار نمی‌پذیرند. کاربران کاملاً مسئول اطمینان از داشتن مجوز مناسب قبل از انجام هرگونه آزمایش امنیتی هستند.

## ویژگی‌ها

- قابلیت فلود HTTP/HTTPS از طریق پروکسی‌ها
- سطح همزمانی قابل تنظیم (تعداد کارگران)
- اعتبارسنجی پروکسی قبل از حمله
- چرخش خودکار پروکسی
- گزارش‌دهی آمار دقیق
- پشتیبانی از خاموشی نرم
- پارامترهای درخواست قابل تنظیم
- گیرنده خودکار پروکسی از منابع عمومی

## نصب

```bash
# کلون کردن مخزن (اگر از گیت استفاده می‌کنید)
git clone https://github.com/bardiavam/goflood.git
cd goflood

# ساخت فایل اجرایی
go build -o goflood

# روش جایگزین: نصب مستقیم با Go
go install github.com/bardiavam/goflood@latest
```

## نحوه استفاده

```bash
# استفاده پایه
./goflood -target http://example.com -proxies proxies.txt -duration 30s

# استفاده پیشرفته با تمام گزینه‌ها
./goflood -target https://example.com -proxies proxies.txt -workers 1000 -duration 2m -proxycheck http://google.com -skipcheck -verbose

# استفاده از گیرنده پروکسی
./goflood -target http://example.com -grabproxies -graboutput my_proxies.txt -duration 1m
```

### پرچم‌های ضروری

- `-target`: آدرس URL هدف برای حمله (باید شامل http:// یا https:// باشد)
- `-proxies` یا `-grabproxies`: یا فایلی با پروکسی‌ها مشخص کنید یا از گیرنده برای دریافت آن‌ها استفاده کنید

### پرچم‌های اختیاری

- `-workers`: تعداد کارگران همزمان (پیش‌فرض: 500)
- `-duration`: مدت حمله (پیش‌فرض: 60s، فرمت: 30s، 5m، 2h)
- `-proxycheck`: URL برای بررسی پروکسی (پیش‌فرض: نقطه پایانی generate_204 گوگل)
- `-skipcheck`: رد کردن مرحله اعتبارسنجی پروکسی (استفاده از تمام پروکسی‌ها صرف نظر از اینکه کار می‌کنند یا نه)
- `-verbose`: فعال‌سازی گزارش‌دهی دقیق در طول حمله

### گزینه‌های گیرنده پروکسی

- `-grabproxies`: فعال‌سازی گرفتن خودکار پروکسی از منابع عمومی
- `-graboutput`: نام فایل برای ذخیره پروکسی‌های گرفته شده (پیش‌فرض: proxies.txt)

## تماس

برای سوالات، پیشنهادات یا مشارکت‌ها، لطفاً تماس بگیرید:

- نام: بردیا
- گیت‌هاب: [@bardiavam](https://github.com/bardiavam)
- ایمیل: bardiawam@gmail.com
