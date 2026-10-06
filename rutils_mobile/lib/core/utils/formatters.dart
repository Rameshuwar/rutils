import 'package:intl/intl.dart';

class Formatters {
  static final NumberFormat _inrFormat = NumberFormat.currency(
    locale: 'en_IN',
    symbol: '₹',
    decimalDigits: 2,
  );

  static final NumberFormat _compactInr = NumberFormat.currency(
    locale: 'en_IN',
    symbol: '₹',
    decimalDigits: 0,
  );

  static String inr(num amount, {bool compact = false}) {
    if (compact) {
      if (amount >= 10000000) {
        return '₹${(amount / 10000000).toStringAsFixed(2)} Cr';
      }
      if (amount >= 100000) {
        return '₹${(amount / 100000).toStringAsFixed(2)} L';
      }
      return _compactInr.format(amount);
    }
    return _inrFormat.format(amount);
  }

  static String number(num value, {int decimals = 2}) {
    if (value % 1 == 0 && decimals <= 2) {
      return value.toInt().toString();
    }
    return value.toStringAsFixed(decimals);
  }

  static String fileSize(int bytes) {
    if (bytes <= 0) return '0 B';
    const suffixes = ['B', 'KB', 'MB', 'GB', 'TB'];
    var i = 0;
    double size = bytes.toDouble();
    while (size >= 1024 && i < suffixes.length - 1) {
      size /= 1024;
      i++;
    }
    return '${size.toStringAsFixed(2)} ${suffixes[i]}';
  }

  static String date(DateTime dt) {
    return DateFormat('yyyy-MM-dd').format(dt);
  }

  static String dateTimeFriendly(DateTime dt) {
    return DateFormat('dd MMM yyyy, hh:mm a').format(dt);
  }
}
