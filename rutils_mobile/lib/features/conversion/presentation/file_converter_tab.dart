import 'dart:typed_data';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/file_download_helper.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/widgets/app_dropdown.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../services/conversion_service.dart';

class FileConverterTab extends StatefulWidget {
  const FileConverterTab({super.key});

  @override
  State<FileConverterTab> createState() => _FileConverterTabState();
}

class _FileConverterTabState extends State<FileConverterTab> {
  final ConversionService _service = ConversionService();

  PlatformFile? _selectedFile;
  Uint8List? _fileBytes;
  int _fileSize = 0;
  String? _detectedFormat;
  String _toFormat = 'pdf';
  bool _isLoading = false;
  String? _statusMessage;
  bool _isSuccess = false;
  Uint8List? _convertedBytes;
  String? _convertedFileName;

  final List<String> _supportedFormats = [
    'pdf',
    'docx',
    'txt',
    'csv',
    'json',
    'jpg',
    'png',
  ];

  Future<void> _pickFile() async {
    try {
      final result = await FilePicker.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['csv', 'json', 'txt', 'docx', 'pdf', 'jpg', 'jpeg', 'png'],
      );

      if (result.isNotEmpty) {
        final file = result.first;
        final bytes = await file.readAsBytes();
        final ext = (file.extension ?? '').toLowerCase();
        final detected = ext == 'jpeg' ? 'jpg' : ext;

        setState(() {
          _selectedFile = file;
          _fileBytes = bytes;
          _fileSize = file.lengthSync() ?? bytes.length;
          _detectedFormat = detected;
          _statusMessage = null;
          _convertedBytes = null;
          _convertedFileName = null;

          if (_toFormat == detected) {
            final alt = _supportedFormats.firstWhere(
              (f) => f != detected,
              orElse: () => 'pdf',
            );
            _toFormat = alt;
          }
        });
      }
    } catch (e) {
      setState(() => _statusMessage = 'Failed to pick file: $e');
    }
  }

  Future<void> _convert() async {
    if (_selectedFile == null || _detectedFormat == null || _fileBytes == null) {
      setState(() {
        _statusMessage = 'Please select a file first.';
        _isSuccess = false;
      });
      return;
    }

    if (_detectedFormat == _toFormat) {
      setState(() {
        _statusMessage = 'Source and target formats cannot be the same.';
        _isSuccess = false;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _statusMessage = 'Converting ${_detectedFormat!.toUpperCase()} to ${_toFormat.toUpperCase()}...';
      _isSuccess = false;
    });

    try {
      final resultBytes = await _service.convertFile(
        fileBytes: _fileBytes!,
        fileName: _selectedFile!.name,
        fromType: _detectedFormat!,
        toType: _toFormat,
      );

      final originalBase = _selectedFile!.name.split('.').first;
      final newFileName = '${originalBase}_converted.$_toFormat';

      setState(() {
        _isLoading = false;
        _isSuccess = true;
        _convertedBytes = Uint8List.fromList(resultBytes);
        _convertedFileName = newFileName;
        _statusMessage = 'Success! Converted to ${_toFormat.toUpperCase()}.';
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _isSuccess = false;
        _statusMessage = e.toString();
      });
    }
  }

  Future<void> _saveOrShare() async {
    if (_convertedBytes == null || _convertedFileName == null) return;
    try {
      await FileDownloadHelper.saveOrShareFile(
        bytes: _convertedBytes!,
        fileName: _convertedFileName!,
      );
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to share/save: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const SectionHeader(
            title: 'File Converter',
            subtitle: 'Convert CSV, JSON, TXT, DOCX, PDF, and Images seamlessly',
            icon: Icons.transform_rounded,
          ),
          const SizedBox(height: 12),

          // Upload Drop Zone / Picker
          GlassCard(
            padding: const EdgeInsets.all(20),
            child: Column(
              children: [
                BouncyButton(
                  onTap: _pickFile,
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(vertical: 28, horizontal: 20),
                    decoration: BoxDecoration(
                      color: const Color(0xFFEFF6FF),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color: const Color(0xFF93C5FD),
                        width: 1.5,
                        strokeAlign: BorderSide.strokeAlignInside,
                      ),
                    ),
                    child: Column(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(14),
                          decoration: const BoxDecoration(
                            color: Colors.white,
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(
                            Icons.cloud_upload_outlined,
                            size: 32,
                            color: AppTheme.primary,
                          ),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          _selectedFile == null ? 'Tap to Select File' : _selectedFile!.name,
                          textAlign: TextAlign.center,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 15,
                            fontWeight: FontWeight.w700,
                            color: AppTheme.textPrimary,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          _selectedFile == null
                              ? 'Supports CSV, JSON, TXT, DOCX, PDF, JPG, PNG'
                              : '${Formatters.fileSize(_fileSize)} • ${_detectedFormat?.toUpperCase()}',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            color: AppTheme.textMuted,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),

                // From & To selectors
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Convert From',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: AppTheme.textSecondary,
                            ),
                          ),
                          const SizedBox(height: 6),
                          Container(
                            height: 48,
                            padding: const EdgeInsets.symmetric(horizontal: 14),
                            decoration: BoxDecoration(
                              color: const Color(0xFFF1F5F9),
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(color: const Color(0xFFE2E8F0)),
                            ),
                            alignment: Alignment.centerLeft,
                            child: Text(
                              _detectedFormat != null ? _detectedFormat!.toUpperCase() : 'Auto-detected',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 13,
                                fontWeight: FontWeight.w700,
                                color: _detectedFormat != null ? AppTheme.primary : AppTheme.textMuted,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 12),
                    const Padding(
                      padding: EdgeInsets.only(top: 24),
                      child: Icon(Icons.arrow_forward_rounded, color: AppTheme.primary, size: 20),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: AppDropdown<String>(
                        label: 'Convert To',
                        value: _toFormat,
                        items: _supportedFormats.map((fmt) {
                          IconData ic;
                          switch (fmt) {
                            case 'pdf':
                              ic = Icons.picture_as_pdf_rounded;
                              break;
                            case 'docx':
                              ic = Icons.description_rounded;
                              break;
                            case 'csv':
                              ic = Icons.table_chart_rounded;
                              break;
                            case 'json':
                              ic = Icons.code_rounded;
                              break;
                            case 'jpg':
                            case 'png':
                              ic = Icons.image_rounded;
                              break;
                            default:
                              ic = Icons.text_snippet_rounded;
                          }
                          return AppDropdownItem(
                            value: fmt,
                            label: fmt.toUpperCase(),
                            subtitle: '$fmt format document/image',
                            icon: ic,
                          );
                        }).toList(),
                        modalTitle: 'Select Target Format',
                        onChanged: (val) => setState(() => _toFormat = val),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
                GradientButton(
                  text: 'Convert Now',
                  isLoading: _isLoading,
                  icon: Icons.sync_rounded,
                  onPressed: _convert,
                ),
              ],
            ),
          ),

          // Status & Result
          if (_statusMessage != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: (_isSuccess ? AppTheme.success : AppTheme.danger).withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(14),
                border: Border.all(
                  color: (_isSuccess ? AppTheme.success : AppTheme.danger).withValues(alpha: 0.3),
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    _isSuccess ? Icons.check_circle_rounded : Icons.error_outline_rounded,
                    color: _isSuccess ? AppTheme.success : AppTheme.danger,
                    size: 20,
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      _statusMessage!,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: _isSuccess ? AppTheme.success : AppTheme.danger,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],

          if (_convertedBytes != null) ...[
            const SizedBox(height: 16),
            GlassCard(
              padding: const EdgeInsets.all(18),
              backgroundColor: const Color(0xFFEFF6FF),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: AppTheme.primary,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.download_done_rounded, color: Colors.white, size: 24),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _convertedFileName ?? 'converted_file',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 14,
                            fontWeight: FontWeight.w700,
                            color: AppTheme.textPrimary,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        Text(
                          Formatters.fileSize(_convertedBytes!.length),
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            color: AppTheme.textMuted,
                          ),
                        ),
                      ],
                    ),
                  ),
                  BouncyButton(
                    onTap: _saveOrShare,
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                      decoration: BoxDecoration(
                        color: AppTheme.primary,
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Text(
                        'Share / Save',
                        style: GoogleFonts.plusJakartaSans(
                          color: Colors.white,
                          fontSize: 12,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}
