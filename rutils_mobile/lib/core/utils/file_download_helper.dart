import 'package:flutter/foundation.dart';
import 'package:path_provider/path_provider.dart';
import 'package:share_plus/share_plus.dart';
import 'dart:io' as io;

class FileDownloadHelper {
  static Future<void> saveOrShareFile({
    required Uint8List bytes,
    required String fileName,
    String? mimeType,
  }) async {
    if (kIsWeb) {
      final xFile = XFile.fromData(
        bytes,
        name: fileName,
        mimeType: mimeType,
      );
      await SharePlus.instance.share(
        ShareParams(
          files: [xFile],
          text: 'Converted file: $fileName',
        ),
      );
      return;
    }

    try {
      final dir = await getTemporaryDirectory();
      final filePath = '${dir.path}/$fileName';
      final file = io.File(filePath);
      await file.writeAsBytes(bytes);

      final xFile = XFile(filePath, name: fileName, mimeType: mimeType);
      await SharePlus.instance.share(
        ShareParams(
          files: [xFile],
          text: 'Converted file: $fileName',
        ),
      );
    } catch (e) {
      debugPrint('Error saving/sharing file: $e');
      rethrow;
    }
  }
}
