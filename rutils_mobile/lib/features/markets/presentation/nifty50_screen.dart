import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/bouncy_button.dart';
import '../../../core/widgets/glass_card.dart';
import '../../../core/widgets/status_badge.dart';
import '../models/nifty_company_model.dart';
import '../services/nifty_service.dart';

class Nifty50Screen extends StatefulWidget {
  const Nifty50Screen({super.key});

  @override
  State<Nifty50Screen> createState() => _Nifty50ScreenState();
}

class _Nifty50ScreenState extends State<Nifty50Screen> {
  final NiftyService _service = NiftyService();
  final _searchController = TextEditingController();

  List<NiftyCompanyModel> _allCompanies = [];
  List<NiftyCompanyModel> _filteredCompanies = [];
  String _selectedIndustry = 'All';
  List<String> _industries = ['All'];
  String? _updatedAt;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadData();
    _searchController.addListener(_filterList);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final res = await _service.fetchCompanies();
      final indSet = <String>{'All'};
      for (final c in res.companies) {
        if (c.industry.isNotEmpty) indSet.add(c.industry);
      }

      setState(() {
        _isLoading = false;
        _allCompanies = res.companies;
        _updatedAt = res.updatedAt;
        _industries = indSet.toList()..sort();
        _filterList();
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
        _errorMessage = e.toString();
      });
    }
  }

  void _filterList() {
    final query = _searchController.text.trim().toLowerCase();
    setState(() {
      _filteredCompanies = _allCompanies.where((c) {
        final matchesQuery = query.isEmpty ||
            c.symbol.toLowerCase().contains(query) ||
            c.companyName.toLowerCase().contains(query) ||
            c.isin.toLowerCase().contains(query);

        final matchesIndustry = _selectedIndustry == 'All' || c.industry == _selectedIndustry;

        return matchesQuery && matchesIndustry;
      }).toList();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.scaffoldBg,
      body: SafeArea(
        child: Column(
          children: [
            // Top Bar
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 16, 20, 12),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      gradient: AppTheme.heroGradient,
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(Icons.trending_up_rounded, color: Colors.white, size: 20),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'NIFTY 50 Constituents',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 18,
                            fontWeight: FontWeight.w800,
                            color: AppTheme.textPrimary,
                          ),
                        ),
                        Text(
                          _updatedAt != null && _updatedAt!.isNotEmpty
                              ? 'Last sync: $_updatedAt'
                              : 'Real-time NSE constituent database',
                          style: GoogleFonts.plusJakartaSans(fontSize: 12, color: AppTheme.textMuted),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  BouncyButton(
                    onTap: _loadData,
                    child: Container(
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(10),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: const Icon(Icons.refresh_rounded, size: 18, color: AppTheme.primary),
                    ),
                  ),
                ],
              ),
            ),

            // Search Bar
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 4),
              child: TextField(
                controller: _searchController,
                decoration: InputDecoration(
                  hintText: 'Search by symbol, company, or ISIN...',
                  prefixIcon: const Icon(Icons.search_rounded, color: AppTheme.primary, size: 20),
                  suffixIcon: _searchController.text.isNotEmpty
                      ? IconButton(
                          icon: const Icon(Icons.clear_rounded, size: 18),
                          onPressed: () => _searchController.clear(),
                        )
                      : null,
                  filled: true,
                  fillColor: Colors.white,
                  contentPadding: const EdgeInsets.symmetric(vertical: 12),
                ),
              ),
            ),
            const SizedBox(height: 8),

            // Industries Filter Horizontal
            if (_industries.length > 1) ...[
              SizedBox(
                height: 36,
                child: ListView.builder(
                  scrollDirection: Axis.horizontal,
                  physics: const BouncingScrollPhysics(),
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  itemCount: _industries.length,
                  itemBuilder: (context, idx) {
                    final ind = _industries[idx];
                    final isSel = ind == _selectedIndustry;
                    return Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: BouncyButton(
                        onTap: () {
                          setState(() {
                            _selectedIndustry = ind;
                            _filterList();
                          });
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 14),
                          decoration: BoxDecoration(
                            color: isSel ? AppTheme.primary : Colors.white,
                            borderRadius: BorderRadius.circular(10),
                            border: Border.all(
                              color: isSel ? AppTheme.primary : const Color(0xFFE2E8F0),
                            ),
                          ),
                          alignment: Alignment.center,
                          child: Text(
                            ind,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: isSel ? Colors.white : AppTheme.textSecondary,
                            ),
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
              const SizedBox(height: 10),
            ],

            // Content List
            Expanded(
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator(color: AppTheme.primary))
                  : _errorMessage != null
                      ? Center(
                          child: Padding(
                            padding: const EdgeInsets.all(24),
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                const Icon(Icons.cloud_off_rounded, size: 48, color: AppTheme.danger),
                                const SizedBox(height: 12),
                                Text(
                                  _errorMessage!,
                                  textAlign: TextAlign.center,
                                  style: GoogleFonts.plusJakartaSans(fontSize: 13, color: AppTheme.danger),
                                ),
                                const SizedBox(height: 16),
                                ElevatedButton.icon(
                                  onPressed: _loadData,
                                  icon: const Icon(Icons.refresh),
                                  label: const Text('Try Again'),
                                ),
                              ],
                            ),
                          ),
                        )
                      : _filteredCompanies.isEmpty
                          ? Center(
                              child: Text(
                                'No matching companies found.',
                                style: GoogleFonts.plusJakartaSans(fontSize: 14, color: AppTheme.textMuted),
                              ),
                            )
                          : RefreshIndicator(
                              onRefresh: _loadData,
                              color: AppTheme.primary,
                              child: ListView.builder(
                                physics: const BouncingScrollPhysics(),
                                padding: const EdgeInsets.fromLTRB(20, 4, 20, 20),
                                itemCount: _filteredCompanies.length,
                                itemBuilder: (context, idx) {
                                  final c = _filteredCompanies[idx];
                                  return Container(
                                    margin: const EdgeInsets.only(bottom: 12),
                                    child: GlassCard(
                                      padding: const EdgeInsets.all(16),
                                      child: Row(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Container(
                                            width: 44,
                                            height: 44,
                                            decoration: BoxDecoration(
                                              gradient: AppTheme.primaryGradient,
                                              borderRadius: BorderRadius.circular(12),
                                            ),
                                            alignment: Alignment.center,
                                            child: Text(
                                              c.symbol.isNotEmpty ? c.symbol.substring(0, 1) : 'S',
                                              style: GoogleFonts.plusJakartaSans(
                                                fontSize: 18,
                                                fontWeight: FontWeight.w800,
                                                color: Colors.white,
                                              ),
                                            ),
                                          ),
                                          const SizedBox(width: 14),
                                          Expanded(
                                            child: Column(
                                              crossAxisAlignment: CrossAxisAlignment.start,
                                              children: [
                                                Row(
                                                  children: [
                                                    Text(
                                                      c.symbol,
                                                      style: GoogleFonts.plusJakartaSans(
                                                        fontSize: 15,
                                                        fontWeight: FontWeight.w800,
                                                        color: AppTheme.primary,
                                                        letterSpacing: -0.2,
                                                      ),
                                                    ),
                                                    const SizedBox(width: 8),
                                                    StatusBadge(label: c.series, color: AppTheme.textMuted),
                                                  ],
                                                ),
                                                const SizedBox(height: 3),
                                                Text(
                                                  c.companyName,
                                                  style: GoogleFonts.plusJakartaSans(
                                                    fontSize: 13,
                                                    fontWeight: FontWeight.w600,
                                                    color: AppTheme.textPrimary,
                                                  ),
                                                ),
                                                const SizedBox(height: 6),
                                                Row(
                                                  children: [
                                                    StatusBadge(label: c.industry, color: const Color(0xFF0284C7)),
                                                    const SizedBox(width: 8),
                                                    Text(
                                                      c.isin,
                                                      style: GoogleFonts.jetBrainsMono(
                                                        fontSize: 11,
                                                        color: AppTheme.textMuted,
                                                      ),
                                                    ),
                                                  ],
                                                ),
                                              ],
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  ).animate().fadeIn(duration: 200.ms);
                                },
                              ),
                            ),
            ),
          ],
        ),
      ),
    );
  }
}
