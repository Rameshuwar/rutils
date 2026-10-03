import 'dart:typed_data';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/file_download_helper.dart';
import '../../../core/utils/formatters.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/custom_text_field.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../services/conversion_service.dart';

class PdfCompressorTab extends StatefulWidget {
  const PdfCompressorTab({super.key});

  @override
  State<PdfCompressorTab> createState() => _PdfCompressorTabState();
}

class _PdfCompressorTabState extends State<PdfCompressorTab> {
  final ConversionService _service = ConversionService();
  final _targetSizeController = TextEditingController(text: '50');

  PlatformFile? _selectedFile;
  Uint8List? _fileBytes;
  int _fileSize = 0;
  String _conversionType = 'compress'; // compress | expand
  String _dataType = 'percentage'; // percentage | size
  bool _isLoading = false;
  String? _statusMessage;
  bool _isSuccess = false;
  Uint8List? _resultBytes;
  String? _resultFileName;

  @override
  void dispose() {
    _targetSizeController.dispose();
    super.dispose();
  }

  Future<void> _pickPdf() async {
    try {
      final res = await FilePicker.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['pdf'],
      );
      if (res.isNotEmpty) {
        final file = res.first;
        final bytes = await file.readAsBytes();
        setState(() {
          _selectedFile = file;
          _fileBytes = bytes;
          _fileSize = bytes.length;
          _statusMessage = null;
          _resultBytes = null;
        });
      }
    } catch (e) {
      setState(() => _statusMessage = 'Failed to select PDF: $e');
    }
  }

  Future<void> _process() async {
    if (_selectedFile == null || _fileBytes == null) {
      setState(() {
        _statusMessage = 'Please select a PDF file first.';
        _isSuccess = false;
      });
      return;
    }

    final targetVal = double.tryParse(_targetSizeController.text);
    if (targetVal == null || targetVal <= 0) {
      setState(() {
        _statusMessage = 'Please enter a valid positive target number.';
        _isSuccess = false;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _statusMessage = 'Processing PDF...';
      _isSuccess = false;
    });

    try {
      final result = await _service.convertPdfSize(
        fileBytes: _fileBytes!,
        fileName: _selectedFile!.name,
        conversionType: _conversionType,
        dataType: _dataType,
        targetSize: targetVal,
      );

      final outName = '${_selectedFile!.name.split('.').first}_$_conversionType.pdf';

      setState(() {
        _isLoading = false;
        _isSuccess = true;
        _resultBytes = Uint8List.fromList(result);
        _resultFileName = outName;
        _statusMessage = 'Success! Processed PDF (${Formatters.fileSize(result.length)}).';
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _isSuccess = false;
        _statusMessage = e.toString();
      });
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
            title: 'PDF Size Optimizer',
            subtitle: 'Compress or expand PDF documents to target size or percentage',
            icon: Icons.picture_as_pdf_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                BouncyButton(
                  onTap: _pickPdf,
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.symmetric(vertical: 24, horizontal: 20),
                    decoration: BoxDecoration(
                      color: const Color(0xFFEFF6FF),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: const Color(0xFF93C5FD)),
                    ),
                    child: Column(
                      children: [
                        const Icon(Icons.picture_as_pdf_outlined, size: 36, color: AppTheme.primary),
                        const SizedBox(height: 10),
                        Text(
                          _selectedFile == null ? 'Select PDF Document' : _selectedFile!.name,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 15,
                            fontWeight: FontWeight.w700,
                            color: AppTheme.textPrimary,
                          ),
                          textAlign: TextAlign.center,
                        ),
                        if (_selectedFile != null)
                          Text(
                            Formatters.fileSize(_fileSize),
                            style: GoogleFonts.plusJakartaSans(fontSize: 12, color: AppTheme.textMuted),
                          ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 20),

                // Conversion Type Selector
                Text(
                  'Optimization Mode',
                  style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600, color: AppTheme.textSecondary),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: BouncyButton(
                        onTap: () => setState(() => _conversionType = 'compress'),
                        child: Container(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          decoration: BoxDecoration(
                            color: _conversionType == 'compress' ? AppTheme.primary : const Color(0xFFF1F5F9),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          alignment: Alignment.center,
                          child: Text(
                            'Compress',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 13,
                              fontWeight: FontWeight.w700,
                              color: _conversionType == 'compress' ? Colors.white : AppTheme.textSecondary,
                            ),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: BouncyButton(
                        onTap: () => setState(() => _conversionType = 'expand'),
                        child: Container(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          decoration: BoxDecoration(
                            color: _conversionType == 'expand' ? AppTheme.primary : const Color(0xFFF1F5F9),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          alignment: Alignment.center,
                          child: Text(
                            'Expand',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 13,
                              fontWeight: FontWeight.w700,
                              color: _conversionType == 'expand' ? Colors.white : AppTheme.textSecondary,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),

                // Data Type Selector
                Text(
                  'Target Unit',
                  style: GoogleFonts.plusJakartaSans(fontSize: 12, fontWeight: FontWeight.w600, color: AppTheme.textSecondary),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: BouncyButton(
                        onTap: () {
                          setState(() {
                            _dataType = 'percentage';
                            _targetSizeController.text = '50';
                          });
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          decoration: BoxDecoration(
                            color: _dataType == 'percentage' ? AppTheme.primaryLight : const Color(0xFFF1F5F9),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          alignment: Alignment.center,
                          child: Text(
                            'Percentage (%)',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 13,
                              fontWeight: FontWeight.w700,
                              color: _dataType == 'percentage' ? Colors.white : AppTheme.textSecondary,
                            ),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: BouncyButton(
                        onTap: () {
                          setState(() {
                            _dataType = 'size';
                            _targetSizeController.text = '200';
                          });
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          decoration: BoxDecoration(
                            color: _dataType == 'size' ? AppTheme.primaryLight : const Color(0xFFF1F5F9),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          alignment: Alignment.center,
                          child: Text(
                            'Target Size (KB)',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 13,
                              fontWeight: FontWeight.w700,
                              color: _dataType == 'size' ? Colors.white : AppTheme.textSecondary,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),

                CustomTextField(
                  controller: _targetSizeController,
                  label: _dataType == 'percentage' ? 'Target Percentage (%)' : 'Target File Size (KB)',
                  hint: _dataType == 'percentage' ? 'e.g. 50' : 'e.g. 200',
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                ),
                const SizedBox(height: 24),

                GradientButton(
                  text: 'Process PDF',
                  isLoading: _isLoading,
                  icon: Icons.tune_rounded,
                  onPressed: _process,
                ),
              ],
            ),
          ),

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

          if (_resultBytes != null) ...[
            const SizedBox(height: 16),
            GlassCard(
              backgroundColor: const Color(0xFFEFF6FF),
              child: Row(
                children: [
                  const Icon(Icons.check_circle_rounded, color: AppTheme.success, size: 28),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _resultFileName ?? 'processed.pdf',
                          style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w700, fontSize: 14),
                        ),
                        Text(
                          'Processed: ${Formatters.fileSize(_resultBytes!.length)}',
                          style: GoogleFonts.plusJakartaSans(color: AppTheme.textMuted, fontSize: 12),
                        ),
                      ],
                    ),
                  ),
                  BouncyButton(
                    onTap: () {
                      FileDownloadHelper.saveOrShareFile(
                        bytes: _resultBytes!,
                        fileName: _resultFileName!,
                      );
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
                      decoration: BoxDecoration(
                        color: AppTheme.primary,
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Text(
                        'Share / Save',
                        style: GoogleFonts.plusJakartaSans(color: Colors.white, fontSize: 12, fontWeight: FontWeight.w700),
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
