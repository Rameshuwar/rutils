import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_dropdown.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/gradient_button.dart';
import '../../../core/widgets/section_header.dart';
import '../../../core/widgets/status_badge.dart';
import '../models/conversion_models.dart';
import '../services/conversion_service.dart';

class TimezoneTab extends StatefulWidget {
  const TimezoneTab({super.key});

  @override
  State<TimezoneTab> createState() => _TimezoneTabState();
}

class _TimezoneTabState extends State<TimezoneTab> {
  final ConversionService _service = ConversionService();

  DateTime _selectedDate = DateTime.now();
  TimeOfDay _selectedTime = TimeOfDay.now();

  static const List<String> _timezones = [
    'UTC',
    'Asia/Kolkata',
    'America/New_York',
    'America/Chicago',
    'America/Denver',
    'America/Los_Angeles',
    'America/Phoenix',
    'America/Sao_Paulo',
    'Europe/London',
    'Europe/Paris',
    'Europe/Berlin',
    'Europe/Rome',
    'Europe/Moscow',
    'Africa/Cairo',
    'Africa/Johannesburg',
    'Africa/Lagos',
    'Asia/Dubai',
    'Asia/Dhaka',
    'Asia/Bangkok',
    'Asia/Singapore',
    'Asia/Hong_Kong',
    'Asia/Tokyo',
    'Australia/Sydney',
    'Australia/Melbourne',
    'Australia/Perth',
    'Pacific/Auckland',
  ];

  String _sourceTz = 'Asia/Kolkata';
  String _destTz = 'America/New_York';
  bool _isLoading = false;
  TimeConversionResult? _result;
  String? _errorMessage;

  Future<void> _pickDate() async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _selectedDate,
      firstDate: DateTime(2000),
      lastDate: DateTime(2050),
    );
    if (picked != null) {
      setState(() => _selectedDate = picked);
    }
  }

  Future<void> _pickTime() async {
    final picked = await showTimePicker(
      context: context,
      initialTime: _selectedTime,
    );
    if (picked != null) {
      setState(() => _selectedTime = picked);
    }
  }

  Future<void> _convert() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await _service.convertTime(
        year: _selectedDate.year,
        month: _selectedDate.month,
        day: _selectedDate.day,
        hour: _selectedTime.hour,
        minute: _selectedTime.minute,
        second: 0,
        sourceTz: _sourceTz,
        destTz: _destTz,
      );

      setState(() {
        _isLoading = false;
        _result = res;
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString();
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
            title: 'Time Zone Converter',
            subtitle: 'Convert dates and times across global timezones with DST accuracy',
            icon: Icons.access_time_rounded,
          ),
          const SizedBox(height: 12),
          GlassCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Date',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: AppTheme.textSecondary,
                            ),
                          ),
                          const SizedBox(height: 6),
                          BouncyButton(
                            onTap: _pickDate,
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: Row(
                                children: [
                                  const Icon(Icons.calendar_today_rounded, size: 16, color: AppTheme.primary),
                                  const SizedBox(width: 8),
                                  Text(
                                    DateFormat('dd MMM yyyy').format(_selectedDate),
                                    style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Time',
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: AppTheme.textSecondary,
                            ),
                          ),
                          const SizedBox(height: 6),
                          BouncyButton(
                            onTap: _pickTime,
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
                              decoration: BoxDecoration(
                                color: Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              child: Row(
                                children: [
                                  const Icon(Icons.schedule_rounded, size: 16, color: AppTheme.primary),
                                  const SizedBox(width: 8),
                                  Text(
                                    _selectedTime.format(context),
                                    style: GoogleFonts.plusJakartaSans(fontSize: 13, fontWeight: FontWeight.w600),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),

                // Source TZ
                AppDropdown<String>(
                  label: 'Source Timezone',
                  value: _sourceTz,
                  items: _timezones.map((tz) {
                    final parts = tz.split('/');
                    final region = parts.length > 1 ? parts[0] : 'Global';
                    final city = (parts.length > 1 ? parts[1] : tz).replaceAll('_', ' ');
                    return AppDropdownItem(
                      value: tz,
                      label: city,
                      subtitle: '$region • $tz',
                      icon: Icons.public_rounded,
                    );
                  }).toList(),
                  modalTitle: 'Select Source Timezone',
                  enableSearch: true,
                  onChanged: (val) => setState(() => _sourceTz = val),
                ),
                const SizedBox(height: 14),

                // Destination TZ
                AppDropdown<String>(
                  label: 'Destination Timezone',
                  value: _destTz,
                  items: _timezones.map((tz) {
                    final parts = tz.split('/');
                    final region = parts.length > 1 ? parts[0] : 'Global';
                    final city = (parts.length > 1 ? parts[1] : tz).replaceAll('_', ' ');
                    return AppDropdownItem(
                      value: tz,
                      label: city,
                      subtitle: '$region • $tz',
                      icon: Icons.flight_land_rounded,
                    );
                  }).toList(),
                  modalTitle: 'Select Destination Timezone',
                  enableSearch: true,
                  onChanged: (val) => setState(() => _destTz = val),
                ),
                const SizedBox(height: 24),

                GradientButton(
                  text: 'Convert Time',
                  isLoading: _isLoading,
                  icon: Icons.sync_alt_rounded,
                  onPressed: _convert,
                ),
              ],
            ),
          ),

          if (_errorMessage != null) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppTheme.danger.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                _errorMessage!,
                style: GoogleFonts.plusJakartaSans(color: AppTheme.danger, fontSize: 13),
              ),
            ),
          ],

          if (_result != null) ...[
            const SizedBox(height: 16),
            GlassCard(
              backgroundColor: const Color(0xFFEFF6FF),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        _destTz,
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 14,
                          fontWeight: FontWeight.w700,
                          color: AppTheme.primary,
                        ),
                      ),
                      if (_result!.isNextDay)
                        const StatusBadge(label: '+1 Day', color: AppTheme.warning)
                      else if (_result!.isPrevDay)
                        const StatusBadge(label: '-1 Day', color: AppTheme.info),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _result!.destTimeLocal,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 24,
                      fontWeight: FontWeight.w800,
                      color: AppTheme.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 12),
                  const Divider(color: Color(0xFFBFDBFE)),
                  const SizedBox(height: 8),
                  Text(
                    'UTC: ${_result!.destTimeUtc}',
                    style: GoogleFonts.plusJakartaSans(fontSize: 12, color: AppTheme.textMuted),
                  ),
                  Text(
                    'Zone Offset: ${_result!.destOffset >= 0 ? '+' : ''}${(_result!.destOffset / 3600).toStringAsFixed(1)} hrs',
                    style: GoogleFonts.plusJakartaSans(fontSize: 12, color: AppTheme.textMuted),
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
