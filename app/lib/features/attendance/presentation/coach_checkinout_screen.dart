import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:geolocator/geolocator.dart';

import '../data/attendance_repository.dart';

class CoachCheckInOutScreen extends ConsumerStatefulWidget {
  const CoachCheckInOutScreen({super.key, required this.classId});
  final String classId;

  @override
  ConsumerState<CoachCheckInOutScreen> createState() => _CoachCheckInOutScreenState();
}

class _CoachCheckInOutScreenState extends ConsumerState<CoachCheckInOutScreen> {
  bool _busy = false;
  String? _statusMessage;

  Future<Position?> _resolvePosition() async {
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (permission == LocationPermission.denied || permission == LocationPermission.deniedForever) {
      setState(() => _statusMessage = 'Location permission is required to check in.');
      return null;
    }
    if (!await Geolocator.isLocationServiceEnabled()) {
      setState(() => _statusMessage = 'Please enable location services.');
      return null;
    }
    return Geolocator.getCurrentPosition(desiredAccuracy: LocationAccuracy.high);
  }

  Future<void> _checkIn() async {
    setState(() {
      _busy = true;
      _statusMessage = null;
    });
    try {
      final position = await _resolvePosition();
      if (position == null) return;
      final result = await ref.read(attendanceRepositoryProvider).checkIn(widget.classId, position.latitude, position.longitude);
      setState(() => _statusMessage = result.checkInVerified
          ? 'Checked in successfully.'
          : 'Checked in, but location could not be verified.');
    } catch (e) {
      setState(() => _statusMessage = _friendlyError(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _checkOut() async {
    setState(() {
      _busy = true;
      _statusMessage = null;
    });
    try {
      final position = await _resolvePosition();
      if (position == null) return;
      await ref.read(attendanceRepositoryProvider).checkOut(widget.classId, position.latitude, position.longitude);
      setState(() => _statusMessage = 'Checked out successfully.');
    } catch (e) {
      setState(() => _statusMessage = _friendlyError(e));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  String _friendlyError(Object e) {
    final text = e.toString();
    if (text.contains('geofence') || text.contains('422')) {
      return 'You are not within the required location radius for this class.';
    }
    return 'Something went wrong. Please try again.';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Class check-in')),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.location_on, size: 72, color: Theme.of(context).colorScheme.primary),
            const SizedBox(height: 16),
            const Text(
              'Your location will be verified against the class location before check-in/out is recorded.',
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.grey),
            ),
            const SizedBox(height: 32),
            if (_statusMessage != null) ...[
              Text(_statusMessage!, textAlign: TextAlign.center),
              const SizedBox(height: 16),
            ],
            if (_busy)
              const CircularProgressIndicator()
            else
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _checkIn,
                      icon: const Icon(Icons.login),
                      label: const Text('Check in'),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed: _checkOut,
                      icon: const Icon(Icons.logout),
                      label: const Text('Check out'),
                    ),
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }
}
